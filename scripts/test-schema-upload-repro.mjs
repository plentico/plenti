// Repro for the reported schema-upload regression:
//   upload display-1800.jpg into a schema field (25x25/crop/webp) ->
//   review modal (schema pre-applied) -> confirm -> field gets the derivative.
// Run: node scripts/test-schema-upload-repro.mjs
import assert from 'node:assert/strict';
import { component, controlledImages, pendingMedia, settle } from './helpers/cms-transition-harness.mjs';

const images = controlledImages();
const originalDocument = globalThis.document;
globalThis.document = {
    createElement: () => ({
        getContext: () => ({ drawImage() {}, fillRect() {} }),
        toBlob: (callback, type) => callback(new Blob(['pixels'], { type })),
    }),
};

const schema = {
    'fields.image.src': {
        type: 'media',
        options: [{ crop: true, scale: true, width: 25, height: 25, convert: 'webp' }],
    },
};

try {
    pendingMedia.clear();
    const field = await component('fields/media.svelte', {
        field: '/media/accessibility.webp', localMediaList: [], changingMedia: '',
        showMediaModal: false, uploadContext: { kind: 'standalone' },
        parentKeys: 'fields.image.src', schema,
    }, `
        openPicker() { swapMedia(); flush(); },
        context() { return uploadContext; },
        save(path) { return onSavedPath(path); },
        read() { flush(); return { field, cropError, imageOptions }; }
    `);
    field.openPicker();
    const context = field.context();
    assert.equal(context.kind, 'field');
    const opts = field.read().imageOptions;
    assert.ok(opts && opts.width === 25 && opts.height === 25 && opts.crop === true && opts.convert === 'webp',
        'imageOptions: ' + JSON.stringify(opts));
    assert.deepEqual(context.imageOptions, opts, 'pick context carries the field schema');

    const view = await component('file_upload.svelte', {
        media: [], localMediaList: [], mediaPrefix: '/', showMediaModal: true,
        uploadContext: context, changingMedia: '',
    }, `
        choose: selectFile,
        next(detail) { return onModalNext({ detail }); },
        close() { onModalClose(); },
        read() { flush(); return { showCropModal, cropError, fieldSchema, processing }; }
    `);

    // 1) Schema field upload must ALWAYS open the review modal (no decode gate).
    view.choose([new File(['original bytes'], 'display-1800.jpg', { type: 'image/jpeg' })]);
    await settle();
    assert.equal(view.read().showCropModal, true, 'schema field upload opens the review modal');
    const fs = view.read().fieldSchema;
    assert.ok(fs && fs.width === 25 && fs.height === 25 && fs.crop === true, 'fieldSchema: ' + JSON.stringify(fs));
    assert.equal(images.loads.length, 0, 'schema route skips the library conformance probe');

    // 2) Confirm with the schema preset's edits -> derivative staged + handed off.
    const image = { naturalWidth: 1800, naturalHeight: 1200 };
    await view.next({
        image,
        selection: { x: 300, y: 0, width: 1200, height: 1200 },
        overrides: { crop: true, width: 25, height: 25, convert: 'webp', maxWidth: null, maxHeight: null },
    });
    await settle();
    assert.equal(view.read().cropError, '');
    const saved = view.events.find(e => e.type === 'saved');
    assert.ok(saved, 'a saved event was dispatched');
    assert.match(saved.detail, /^\/media\/display-1800-[0-9a-f]+-25x25\.webp$/, 'derivative path: ' + saved.detail);
    assert.ok(pendingMedia.previewUrl(saved.detail), 'derivative staged in pendingMedia');

    // 3) The field accepts the handoff and keeps the derivative path.
    const p3 = field.save(saved.detail);
    await settle();
    assert.equal(images.loads.length, 1, 'field probed via the pendingMedia preview');
    images.resolve(images.loads[0], 25, 25);
    await p3; await settle();
    assert.equal(field.read().cropError, '');
    assert.equal(field.read().field, saved.detail);

    console.log('PASS schema-field upload confirm chain');

    // 4) Close (x / backdrop) without confirming -> CANCELS the swap: nothing
    //    is staged, no path is handed to the field, and the field value and the
    //    pending queue stay untouched.
    pendingMedia.clear();
    const savedBefore = view.events.filter(e => e.type === 'saved').length;
    view.choose([new File(['raw bytes'], 'raw.jpg', { type: 'image/jpeg' })]);
    await settle();
    assert.equal(view.read().showCropModal, true, 'review modal opens again');
    view.close();
    await settle();
    assert.equal(view.read().showCropModal, false, 'close dismisses the review modal');
    assert.equal(view.events.filter(e => e.type === 'saved').length, savedBefore,
        'cancel stages nothing and hands no path back');
    assert.equal(field.read().field, saved.detail,
        'field keeps its previous value: ' + field.read().field);
    assert.equal(field.read().cropError, '');
    console.log('PASS modal close cancels the swap (field untouched)');

    // 5) Reset + confirm (untouched) -> keep the original, no schema re-apply.
    pendingMedia.clear();
    view.choose([new File(['raw bytes'], 'keep.jpg', { type: 'image/jpeg' })]);
    await settle();
    await view.next({ image, selection: null, overrides: null });
    await settle();
    const keepSaved = view.events.filter(e => e.type === 'saved').at(-1);
    assert.equal(keepSaved.detail, '/media/keep.jpg');
    assert.equal(context.bypassSchema, true, 'untouched confirm flags bypassSchema');
    await field.save(keepSaved.detail);
    await settle();
    assert.equal(field.read().field, '/media/keep.jpg', 'field keeps the original');
    assert.equal(context.bypassSchema, false, 'flag is one-shot');
    console.log('PASS untouched confirm keeps the original');
} finally {
    images.restore();
    globalThis.document = originalDocument;
    pendingMedia.clear();
}
