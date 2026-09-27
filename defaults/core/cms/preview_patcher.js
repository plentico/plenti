import { pendingMedia } from './pending_media.js';

// Keeps the LIVE PAGE preview truthful while media is deferred (#364 D13):
// a content edit puts a derivative PATH into the page markup immediately, but
// the file only reaches media/ at the page-save commit — so the page's own
// <img>/<embed> elements would render broken until then. While an entry is
// pending, matching elements are swapped to its in-memory object URL; the
// original path is remembered on the element and restored when the entry
// leaves the store. After a successful save the entry survives markCommitted()
// with its object URL, so the blob keeps covering the just-written file until
// reload (the same freshly-written-path race pending_media.js documents).
//
// The field widget (media.svelte displaySrc) does its own pendingMedia lookup;
// this module covers markup the CMS does not own — the user's page layout.
const ATTR = 'data-plenti-pending-src';

// Content stores bare 'media/…' paths; layouts may render them with a leading
// slash or baseurl prefix, so match on the exact value or a '/'-suffix.
const entryFor = (list, src) =>
    src ? list.find(i => src === i.file || src.endsWith('/' + i.file)) : null;

function restore(el, state) {
    // A layout may have changed src before the observer ran (including during
    // teardown). Only undo the URL WE wrote; never undo the layout's new value.
    if (el.getAttribute('src') === state.url) el.setAttribute('src', state.original);
    el.removeAttribute(ATTR);
}

function applyPending(list, patched) {
    document.querySelectorAll('img, embed').forEach(el => {
        const src = el.getAttribute('src') || '';
        let state = patched.get(el);
        if (state && src !== state.url && src !== state.original) {
            // A different src belongs to a NEW selection, not a rerender of
            // the previous one. Forget A before resolving B (pending or saved).
            restore(el, state);
            patched.delete(el);
            state = null;
        }
        const original = state ? state.original : src;
        const entry = entryFor(list, original);
        if (entry) {
            // Remember the last written URL separately: re-cropping the same
            // path replaces its blob, while its canonical identity stays put.
            patched.set(el, { original, url: entry.url });
            el.setAttribute(ATTR, original);
            if (src !== entry.url) el.setAttribute('src', entry.url);
        } else if (state) {
            restore(el, state);
            patched.delete(el);
        }
    });
    for (const [el, state] of patched) {
        if (!el.isConnected) { restore(el, state); patched.delete(el); }
    }
}

// Client-only: call from onMount. Returns a stop() that restores every
// original src. Re-entry settles in one pass — a patched element's src equals
// the entry URL, so the observer callback makes no further mutations.
export function startPreviewPatcher() {
    let current = [];
    const patched = new Map();
    const run = () => applyPending(current, patched);
    const unsubscribe = pendingMedia.subscribe(list => { current = list; run(); });
    const observer = new MutationObserver(run);
    observer.observe(document.body, {
        childList: true,
        subtree: true,
        attributes: true,
        attributeFilter: ['src'],
    });
    return () => {
        observer.disconnect();
        unsubscribe();
        for (const [el, state] of patched) restore(el, state);
        patched.clear();
    };
}
