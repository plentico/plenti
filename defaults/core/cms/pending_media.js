import { writable, get } from 'svelte/store';
import { blobToDataURL } from './crop-engine.js';

// Image derivatives from the crop engine, held IN MEMORY as Blobs (deferred
// persistence). Nothing is written until the page-save commit; cancelling the
// edit discards these. Blobs (not data URLs) are kept so the larger base64 isn't
// carried through UI state — conversion happens only in toCommitItems().
//
// Each entry keeps an object URL for in-editor preview and remembers the
// ORIGINAL sourcePath so re-cropping transforms the original, not the
// already-compressed derivative.
//
// Two lifecycles share one store:
//   - commit:  toCommitItems() returns entries not yet committed; a successful
//              page save calls markCommitted() so they aren't re-committed.
//   - preview: entries (and their object URLs / sourcePath) survive markCommitted
//              so the thumbnail keeps showing the just-saved derivative until the
//              page reloads (Plenti serves the written file from media/, but the
//              freshly-written path can race the <img> at save time). clear()
//              ends the editing session (page change / cancel): revoke + drop.
//
// PROVENANCE BOUNDARY: sourceOf() spans the current editing session only. After
// a reload the field holds just the derivative path, so a later re-crop uses
// that derivative. Persistent cross-session provenance is future work.
function createPendingMedia() {
    const store = writable([]); // [{ file, blob, sourcePath, url, committed }]
    const { subscribe, update } = store;

    const revoke = item => { if (item && item.url) URL.revokeObjectURL(item.url); };

    return {
        subscribe,
        // Add or REPLACE a pending asset, deduped by output path: re-cropping the
        // same target replaces the blob (revoking the previous preview URL) and
        // keeps the original sourcePath. `action` defaults to the provider-neutral
        // 'upsert' (derivatives may legitimately re-derive an existing path);
        // deferred RAW passthrough files pass 'create' so Gitea/GitLab surface
        // same-name conflicts. The existing local dev endpoint still overwrites.
        add(file, blob, sourcePath, action = 'upsert') {
            update(list => {
                const existing = list.find(i => i.file === file);
                if (existing) revoke(existing);
                return [
                    ...list.filter(i => i.file !== file),
                    { file, blob, sourcePath, action, url: URL.createObjectURL(blob), committed: false },
                ];
            });
        },
        // Standalone uploads already persisted. Retain their own preview before
        // the upload queue is torn down; never mark unrelated page edits saved.
        rememberSaved(file, contents) {
            const comma = contents.indexOf(',');
            const mime = contents.slice(5, contents.indexOf(';'));
            const bytes = Uint8Array.from(atob(contents.slice(comma + 1)), c => c.charCodeAt(0));
            const blob = new Blob([bytes], { type: mime });
            update(list => {
                const existing = list.find(i => i.file === file);
                // A page may hold a newer, unsaved replacement at this path.
                if (existing && !existing.committed) return list;
                revoke(existing);
                return [...list.filter(i => i.file !== file),
                    { file, blob, sourcePath: file, url: URL.createObjectURL(blob), committed: true }];
            });
        },
        remove(file) {
            update(list => {
                const item = list.find(i => i.file === file);
                if (item) revoke(item);
                return list.filter(i => i.file !== file);
            });
        },
        // After a successful page save: stop re-committing these, but keep their
        // previews so the thumbnail stays correct until the page reloads.
        markCommitted() {
            update(list => list.map(i => ({ ...i, committed: true })));
        },
        // End of the editing session (page change / cancel): revoke + drop all.
        clear() {
            update(list => { list.forEach(revoke); return []; });
        },
        // Original source for a derivative path (session-scoped).
        sourceOf(file) {
            const item = get(store).find(i => i.file === file);
            return item ? item.sourcePath : null;
        },
        // Object URL for previewing a derivative that may not be reliably served
        // from media/ yet (this session).
        previewUrl(file) {
            const item = get(store).find(i => i.file === file);
            return item ? item.url : null;
        },
        // Commit items for the page-save payload — only entries not yet committed.
        // Blob -> base64 data URL HERE; each carries its own action/encoding so it
        // merges into the content commit without overriding the content item.
        //
        // action 'upsert' is a provider-neutral intent: a derivative path may or may
        // not already exist (a re-crop in a LATER session re-derives the same
        // filename). Each provider resolves it to create-or-update against the live
        // repo — a fixed 'create' collides cross-session (GitLab rejects the atomic
        // commit, Gitea returns 422). 'upsert' is never sent to a remote API verbatim:
        // the GitLab/Gitea providers translate it (the local provider overwrites, so
        // it treats upsert as create).
        async toCommitItems() {
            return Promise.all(get(store).filter(i => !i.committed).map(async i => ({
                action: i.action ?? 'upsert',
                encoding: 'base64',
                file: i.file,
                contents: await blobToDataURL(i.blob),
            })));
        },
    };
}

export const pendingMedia = createPendingMedia();
