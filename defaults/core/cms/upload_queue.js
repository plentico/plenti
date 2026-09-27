// upload_queue.js — the standalone Media-Library multi-file upload queue.
//
// A small deterministic state machine for a batch of selected files. It owns
// ONLY queue/lifecycle state — no DOM, no provider calls, no transforms, and
// NO payloads (staged transports live in the session's keyed store, owned by
// media_modal). The session owner drives it (claim, advance, resolve, fail,
// skip) and reads it (current, canSave, unresolvedCount).
//
// Per-item states:
//   queued -> preparing -> awaiting_decision -> processing -> resolved
// Terminal: failed, cancelled.
//
// File TYPE is a classification, not a state:
//   'image'       — JPEG/PNG/WebP/AVIF (canvas-processable: optimise/crop)
//   'passthrough' — PDF/SVG/GIF (byte-preserving: original bytes + extension)
//
// Run ownership is SESSION-WIDE, not component-local: an advance/confirm chain
// claims the item it drives via claimRun(item) -> token. The claim is stored
// WITH its owner ({ itemId, token }) so a claim for one item is never mistaken
// for a claim on another. A later claimRun — from ANY component instance —
// invalidates the earlier chain (this is what makes a FileUpload remount safe);
// terminal transitions clear the claim. Chains re-check ownsRun() after every
// await: late completions are IGNORED, not physically aborted.
//
// Invariants the state machine guarantees (the session owner honours the rest):
//   - deterministic order (insertion order, never reordered)
//   - one active item at a time (`current` + the single active run claim)
//   - Save disabled while ANY item is unresolved (queued/preparing/
//     awaiting_decision/processing, or failed until removed)
//   - skipRemaining() retains ONLY resolved items — pending AND failed are
//     dropped, so approved work stays savable and a failed item cannot brick it
//   - a cancelled/foreign item's late result is rejected (state + membership)
//   - object-URL revocation is idempotent (revoke() clears the ref)

const IMAGE_MIMES = ['image/jpeg', 'image/png', 'image/webp', 'image/avif'];
const IMAGE_EXTS = ['jpg', 'jpeg', 'png', 'webp', 'avif'];

// Same regex as crop-engine.js's exported sourceExtension — duplicated ONLY so
// this module stays import-free for the data:-URL test loader. Keep in sync.
function extOf(name) {
    const m = String(name || '').toLowerCase().match(/\.([a-z0-9]+)(?:[?#].*)?$/);
    return m ? m[1] : '';
}

/**
 * Classify a File: 'image' (canvas-processable) or 'passthrough' (byte-preserving).
 * MIME first; extension fallback ONLY when MIME is absent/generic — a real non-image
 * MIME (e.g. application/pdf named "x.jpg") is NOT treated as an image.
 * This is the SINGLE classification source — file_upload.svelte imports it too.
 */
export function classifyFile(file) {
    const mime = String(file?.type || '').toLowerCase();
    const ext = extOf(file?.name);
    if (IMAGE_MIMES.includes(mime)) return 'image';
    if (mime && mime !== 'application/octet-stream') return 'passthrough';
    return IMAGE_EXTS.includes(ext) ? 'image' : 'passthrough';
}

// States that mean "not yet done" — Save must stay disabled while any item is
// here. 'failed' is unresolved too: it blocks Save until removed (the UI wires
// remove() to a per-item control) or the batch is skipped/discarded.
const UNRESOLVED = new Set(['queued', 'preparing', 'awaiting_decision', 'processing', 'failed']);
// States an explicit "Skip remaining" acts on (everything not already approved
// or terminal-cancelled). Also drives the Skip-remaining affordance count.
const SKIPPABLE = new Set(['queued', 'preparing', 'awaiting_decision', 'processing', 'failed']);

let _seq = 0; // monotonic id source (module-local; fine for a browser session)

/**
 * Build a queue from a FileList/array. Each item:
 *   { id, file, type, state, objectUrl, error }
 * (No payloads here — transports are keyed by item.id in the session store.)
 * `makeObjectUrl` is injected (URL.createObjectURL) so the module stays DOM-free.
 */
export function createUploadQueue(files, makeObjectUrl) {
    const items = Array.from(files || []).map(file => ({
        id: ++_seq,
        file,
        type: classifyFile(file),
        state: 'queued',
        objectUrl: null,
        error: null,
    }));

    // The single active run claim: { itemId, token } or null. Stored with its
    // owner so hasActiveRun(other) is never confused by a claim on this item.
    let activeRun = null;
    const clearClaimFor = (item) => {
        if (item && activeRun && activeRun.itemId === item.id) activeRun = null;
    };

    const idx = () => items.findIndex(i => i.state !== 'resolved'
        && i.state !== 'cancelled' && i.state !== 'failed');

    const revoke = (item) => {
        if (item && item.objectUrl) {
            try { URL.revokeObjectURL(item.objectUrl); } catch (_) { /* idempotent */ }
            item.objectUrl = null;
        }
    };

    const member = (item) => !!item && items.includes(item);

    return {
        items,
        // The next item to work, or null when the queue is drained of actionable items.
        get current() {
            const i = idx();
            return i === -1 ? null : items[i];
        },

        // ── run ownership (session-wide; survives component remounts) ────────
        // Claim the right to drive `item`. A later claim invalidates any earlier
        // chain's token, whichever component instance created it.
        claimRun(item) {
            if (!member(item)) return null;
            const token = {};
            activeRun = { itemId: item.id, token };
            return token;
        },
        // Does an unexpired claim exist for THIS item (whoever holds it)?
        hasActiveRun(item) {
            return !!item && !!activeRun && activeRun.itemId === item.id;
        },
        // Is this chain (token) the current claimant for this item?
        ownsRun(item, token) {
            return !!item && !!token && !!activeRun
                && activeRun.itemId === item.id && activeRun.token === token;
        },
        // Release only your own claim (a stolen/expired token is a no-op).
        clearRun(item, token) {
            if (this.ownsRun(item, token)) activeRun = null;
        },

        // ── transitions (membership-guarded; callers check the boolean) ──────
        prepare(item) {
            if (!member(item)) return null;
            item.state = 'preparing';
            item.objectUrl = makeObjectUrl ? makeObjectUrl(item.file) : null;
            return item;
        },
        awaitDecision(item) { if (!member(item)) return null; item.state = 'awaiting_decision'; return item; },
        process(item) { if (!member(item)) return null; item.state = 'processing'; return item; },
        // Approve an item. Rejected for cancelled/foreign items (guards the
        // "late result must not insert" invariant). Payload staging is the
        // session owner's atomic completeItem(); this only transitions state.
        resolve(item) {
            if (!member(item) || item.state === 'cancelled') return false;
            item.state = 'resolved';
            revoke(item);
            clearClaimFor(item);
            return true;
        },
        fail(item, error) {
            if (!member(item) || item.state === 'cancelled') return false;
            item.error = error;
            item.state = 'failed';
            revoke(item);
            clearClaimFor(item);
            return true;
        },
        // Skip the CURRENT item ("Skip this file"): mark cancelled + revoke, so
        // any late result is dropped by resolve()'s state guard.
        cancelCurrent(item) {
            if (!member(item)) return;
            item.state = 'cancelled';
            revoke(item);
            clearClaimFor(item);
        },
        remove(item) {
            if (!member(item)) return;
            revoke(item);
            clearClaimFor(item);
            const i = items.indexOf(item);
            if (i !== -1) items.splice(i, 1);
        },
        // "Skip remaining": retain ONLY resolved items; everything else —
        // pending states AND failed — is cancelled + revoked. Approved work
        // stays savable; a failed item cannot go on blocking Save through this
        // path. Clears any active run claim (its item is no longer driveable).
        skipRemaining() {
            for (const item of items) {
                if (item.state !== 'resolved' && item.state !== 'cancelled') {
                    item.state = 'cancelled';
                }
                if (item.state === 'cancelled') revoke(item);
            }
            activeRun = null;
        },
        // Idempotent teardown — revoke every URL, clear the claim (safe to call
        // more than once; resolved items' URLs were already revoked at resolve).
        destroy() {
            for (const item of items) revoke(item);
            activeRun = null;
        },

        // ── derived state ─────────────────────────────────────────────────────
        // Save may proceed only when NOTHING is unresolved AND ≥1 item resolved.
        get canSave() {
            const anyUnresolved = items.some(i => UNRESOLVED.has(i.state));
            const anyResolved = items.some(i => i.state === 'resolved');
            return !anyUnresolved && anyResolved;
        },
        get unresolvedCount() { return items.filter(i => UNRESOLVED.has(i.state)).length; },
        // Items an explicit skip would act on (drives the Skip-remaining affordance).
        get skippableCount() { return items.filter(i => SKIPPABLE.has(i.state)).length; },
        failedItems() { return items.filter(i => i.state === 'failed'); },
        resolvedIds() { return items.filter(i => i.state === 'resolved').map(i => i.id); },
    };
}
