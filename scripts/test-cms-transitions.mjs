// Run: node scripts/test-cms-transitions.mjs
// Or: deno run --allow-read --allow-env=TEST scripts/test-cms-transitions.mjs
// Synthetic component-transition evidence; see the harness's scope note.
import assert from 'node:assert/strict';
import { Buffer } from 'node:buffer';
import { component, controlledImages, pendingMedia, settle, previewPage, scheduledMedia } from './helpers/cms-transition-harness.mjs';

let passed = 0, failed = 0;
async function test(name, run) {
    const images = controlledImages();
    const originalDocument = globalThis.document;
    const originalReader = globalThis.FileReader;
    pendingMedia.clear();
    try { await run(images); passed++; console.log('PASS ' + name); }
    catch (error) { failed++; console.error('FAIL ' + name + '\n' + error.stack); }
    finally { images.restore(); globalThis.document = originalDocument; globalThis.FileReader = originalReader; pendingMedia.clear(); }
}

const mediaField = (props = {}) => component('fields/media.svelte', {
    field: 'media/original.webp', localMediaList: [], changingMedia: '',
    showMediaModal: false, uploadContext: { kind: 'standalone' }, parentKeys: 'image',
    schema: { image: { type: 'media', options: [{ crop: false, scale: false }] } },
    ...props,
}, `
    pick(path) { swapMedia(); flush(); changingMedia = path; flush(); },
    reopen() { swapMedia(); flush(); },
    read() { flush(); return { field, cropError, showCropModal, processing }; }
`);

await test('latest field selection wins when image A decodes after B', async images => {
    const field = await mediaField();
    field.pick('media/a.webp');
    field.pick('media/b.webp');
    images.resolve(images.loads[1]); await settle();
    assert.equal(field.read().field, 'media/b.webp');
    images.resolve(images.loads[0]); await settle();
    assert.equal(field.read().field, 'media/b.webp');
    field.destroy();
});

await test('reopening the picker invalidates an unfinished selection', async images => {
    const field = await mediaField();
    field.pick('media/a.webp');
    field.reopen();
    images.resolve(images.loads[0]); await settle();
    assert.equal(field.read().field, 'media/original.webp');
    field.destroy();
});

await test('destroyed field ignores its unfinished selection', async images => {
    const field = await mediaField();
    field.pick('media/a.webp');
    field.destroy();
    images.resolve(images.loads[0]); await settle();
    assert.equal(field.read().field, 'media/original.webp');
});

const upload = () => component('file_upload.svelte', {
    media: [], localMediaList: [], mediaPrefix: '', showMediaModal: true,
    uploadContext: { kind: 'field' }, changingMedia: '',
}, `choose: selectFile, read() { return { showCropModal, cropError }; }`);
const stagedPaths = () => {
    let list; pendingMedia.subscribe(value => list = value)();
    return list.map(item => item.file);
};

await test('closing upload while decoding stages nothing and emits no handoff', async images => {
    const view = await upload();
    view.choose([new File(['image'], 'a.webp', { type: 'image/webp' })]);
    view.destroy();
    images.resolve(images.loads[0]); await settle();
    assert.deepEqual({ paths: stagedPaths(), events: view.events }, { paths: [], events: [] });
});

await test('changing upload owner during decode cannot hand A to field B', async images => {
    const view = await upload();
    view.choose([new File(['image'], 'a.webp', { type: 'image/webp' })]);
    view.set({ uploadContext: { kind: 'field' } });
    images.resolve(images.loads[0]); await settle();
    assert.deepEqual({ paths: stagedPaths(), events: view.events }, { paths: [], events: [] });
    view.destroy();
});

await test('replacing a pending image upload with a PDF drops the late image', async images => {
    const view = await upload();
    view.choose([new File(['image'], 'a.webp', { type: 'image/webp' })]);
    view.choose([new File(['pdf'], 'b.pdf', { type: 'application/pdf' })]);
    images.resolve(images.loads[0]); await settle();
    assert.deepEqual(stagedPaths(), ['media/b.pdf']);
    assert.deepEqual(view.events, [{ type: 'saved', detail: 'media/b.pdf' }]);
    view.destroy();
});

await test('latest selection also wins while the older image is being encoded', async images => {
    let finishEncoding;
    globalThis.document = { createElement: () => ({
        getContext: () => ({ drawImage() {}, fillRect() {} }),
        toBlob: (callback, type) => { finishEncoding = () => callback(new Blob(['pixels'], { type })); },
    }) };
    const field = await mediaField({
        schema: { image: { type: 'media', options: [{ crop: false, scale: false, convert: 'webp' }] } },
    });
    field.pick('media/a.png');
    images.resolve(images.loads[0]); await settle();
    assert.equal(typeof finishEncoding, 'function');
    field.pick('media/b.webp');
    images.resolve(images.loads[1]); await settle();
    finishEncoding(); await settle();
    assert.equal(field.read().field, 'media/b.webp');
    assert.deepEqual(stagedPaths(), []);
    assert.equal(field.read().processing, false);
    field.destroy();
});

await test('an active conforming upload stages the original file and hands off its path', async images => {
    const view = await upload();
    const file = new File(['original bytes'], 'active.webp', { type: 'image/webp' });
    view.choose([file]);
    images.resolve(images.loads[0]); await settle();
    let entries; pendingMedia.subscribe(value => entries = value)();
    assert.equal(entries[0].blob, file);
    assert.equal(entries[0].action, 'create');
    assert.deepEqual(view.events, [{ type: 'saved', detail: 'media/active.webp' }]);
    view.destroy();
});

await test('an active nonconforming upload still opens the optimise modal', async images => {
    const view = await upload();
    view.choose([new File(['image'], 'active.png', { type: 'image/png' })]);
    images.resolve(images.loads[0]); await settle();
    assert.equal(view.read().showCropModal, true);
    assert.deepEqual(stagedPaths(), []);
    view.destroy();
});

await test('an older decode failure cannot put an error on the newer selection', async images => {
    const field = await mediaField();
    field.pick('media/a.webp'); field.pick('media/b.webp');
    images.resolve(images.loads[1]); await settle();
    images.reject(images.loads[0]); await settle();
    assert.deepEqual([field.read().field, field.read().cropError], ['media/b.webp', '']);
    field.destroy();
});

await test('page preview follows a new persisted selection instead of retaining A', async () => {
    const page = await previewPage();
    try {
        pendingMedia.add('media/a.webp', new Blob(['A']), 'media/a.webp');
        const hero = page.element('media/a.webp'); page.start();
        assert.equal(hero.getAttribute('src'), pendingMedia.previewUrl('media/a.webp'));
        hero.setAttribute('src', 'media/b.webp'); await settle();
        assert.equal(hero.getAttribute('src'), 'media/b.webp');
        assert.equal(hero.getAttribute('data-plenti-pending-src'), null);
    } finally { page.restore(); }
});

for (const cleanup of ['entry removal', 'patcher stop']) {
    await test('A to pending B restores B on ' + cleanup, async () => {
        const page = await previewPage();
        try {
            pendingMedia.add('media/a.webp', new Blob(['A']), 'media/a.webp');
            pendingMedia.add('media/b.webp', new Blob(['B']), 'media/b.webp');
            const hero = page.element('media/a.webp'); page.start();
            hero.setAttribute('src', 'media/b.webp'); await settle();
            assert.equal(hero.getAttribute('src'), pendingMedia.previewUrl('media/b.webp'));
            if (cleanup === 'entry removal') pendingMedia.remove('media/b.webp');
            else page.stop();
            await settle();
            assert.equal(hero.getAttribute('src'), 'media/b.webp');
            assert.equal(hero.getAttribute('data-plenti-pending-src'), null);
        } finally { page.restore(); }
    });
}

await test('same-path rerender and re-crop keep the current blob and restore the path', async () => {
    const page = await previewPage();
    try {
        pendingMedia.add('media/a.webp', new Blob(['A']), 'media/a.webp');
        const hero = page.element('/base/media/a.webp'); page.start();
        hero.setAttribute('src', '/base/media/a.webp'); await settle();
        assert.equal(hero.getAttribute('src'), pendingMedia.previewUrl('media/a.webp'));
        pendingMedia.add('media/a.webp', new Blob(['new A']), 'media/a.webp'); await settle();
        assert.equal(hero.getAttribute('src'), pendingMedia.previewUrl('media/a.webp'));
        page.stop();
        assert.equal(hero.getAttribute('src'), '/base/media/a.webp');
    } finally { page.restore(); }
});

await test('stopping before a queued observer runs does not restore stale A', async () => {
    const page = await previewPage();
    try {
        pendingMedia.add('media/a.webp', new Blob(['A']), 'media/a.webp');
        const hero = page.element('media/a.webp'); page.start();
        hero.setAttribute('src', 'media/b.webp'); page.stop(); await settle();
        assert.equal(hero.getAttribute('src'), 'media/b.webp');
    } finally { page.restore(); }
});

await test('PDF blur selection returns its canonical tile path, not its preview URL', async () => {
    const originalWindow = globalThis.window, originalFocus = globalThis.focus;
    let blur;
    const attributes = new Map([
        ['src', 'blob:pdf-preview'], ['data-plenti-media-path', 'media/document.pdf'],
    ]);
    const embed = {
        getAttribute: name => attributes.get(name) ?? null,
        attributes: { src: { nodeValue: 'blob:pdf-preview' } },
    };
    globalThis.document = { querySelectorAll: () => [embed], activeElement: embed };
    globalThis.focus = () => {};
    globalThis.window = { parent: { focus() {} }, addEventListener: (_type, callback) => { blur = callback; } };
    try {
        const grid = await component('media_grid.svelte', {
            files: ['media/document.pdf'], changingMedia: 'media/original.webp', showMediaModal: true,
        }, 'read() { return { changingMedia, selectedMedia, showMediaModal }; }');
        grid.mount(); blur();
        assert.deepEqual(grid.read(), {
            changingMedia: 'media/document.pdf', selectedMedia: ['media/document.pdf'], showMediaModal: false,
        });
    } finally { globalThis.window = originalWindow; globalThis.focus = originalFocus; }
});

const librarySession = async (baseurl = '/') => {
    const admin = await component('admin_menu.svelte', {}, `
        read() { flush(); return [...media]; },
        libraryChanged(next) { media = next; flush(); }
    `, {
        '../../generated/media.js': { default: ['media/existing.webp'] },
        '../../generated/env.js': { env: { baseurl } },
    });
    const library = await component('media_browser.svelte', {
        media: admin.read(), changingMedia: '', showMediaModal: true,
    }, `
        select(paths) { selectedMedia = paths; flush(); },
        completeDelete() { removeMedia(); flush(); },
        read() { return [...media]; }
    `);
    return {
        read: () => admin.read(),
        select(paths) { library.set({ media: admin.read() }); library.select(paths); },
        completeDelete() {
            library.completeDelete();
            // The media_browser -> media_modal -> admin_menu bind:media handoff.
            admin.libraryChanged(library.read());
        },
    };
};

await test('successfully deleted saved upload stays absent from the session Library', async () => {
    const library = await librarySession();
    pendingMedia.add('media/new.webp', new Blob(['new']), 'media/new.webp');
    pendingMedia.markCommitted();
    assert.deepEqual(library.read(), ['media/existing.webp', 'media/new.webp']);
    library.select(['media/new.webp']);
    library.completeDelete();
    assert.deepEqual(library.read(), ['media/existing.webp']);
});

await test('deleting a slash-prefixed Library path retires its canonical preview', async () => {
    const library = await librarySession('');
    pendingMedia.add('media/new.webp', new Blob(['new']), 'media/new.webp');
    pendingMedia.markCommitted();
    assert.deepEqual(library.read(), ['/media/existing.webp', '/media/new.webp']);
    library.select(['/media/new.webp']);
    library.completeDelete();
    assert.deepEqual(library.read(), ['/media/existing.webp']);
    assert.equal(pendingMedia.previewUrl('media/new.webp'), null);
});

await test('save then delete then save the same path appends the new asset once', async () => {
    const library = await librarySession();
    const file = 'media/reused.webp';
    pendingMedia.add(file, new Blob(['first']), file);
    pendingMedia.markCommitted();
    library.select([file]); library.completeDelete();
    assert.deepEqual(library.read(), ['media/existing.webp']);
    assert.equal(pendingMedia.previewUrl(file), null);
    pendingMedia.add(file, new Blob(['replacement']), file);
    assert.deepEqual(library.read(), ['media/existing.webp']);
    pendingMedia.markCommitted();
    assert.deepEqual(library.read(), ['media/existing.webp', file]);
    pendingMedia.markCommitted();
    assert.deepEqual(library.read(), ['media/existing.webp', file]);
});

await test('an unrelated later page save cannot resurrect a deleted path', async () => {
    const library = await librarySession();
    pendingMedia.add('media/deleted.webp', new Blob(['deleted']), 'media/deleted.webp');
    pendingMedia.markCommitted();
    library.select(['media/deleted.webp']); library.completeDelete();
    pendingMedia.add('media/later.webp', new Blob(['later']), 'media/later.webp');
    pendingMedia.markCommitted();
    assert.deepEqual(library.read(), ['media/existing.webp', 'media/later.webp']);
});

await test('selecting for deletion without provider success preserves the saved preview', async () => {
    const library = await librarySession();
    pendingMedia.add('media/keep.webp', new Blob(['keep']), 'media/keep.webp');
    pendingMedia.markCommitted();
    const preview = pendingMedia.previewUrl('media/keep.webp');
    library.select(['media/keep.webp']);
    // A cancelled/failed delete never invokes the success-only completion hook.
    assert.deepEqual(library.read(), ['media/existing.webp', 'media/keep.webp']);
    assert.equal(pendingMedia.previewUrl('media/keep.webp'), preview);
});

await test('deleting the persisted path preserves an unsaved replacement for the next save', async () => {
    const library = await librarySession();
    const file = 'media/replacement.webp';
    pendingMedia.add(file, new Blob(['saved']), file); pendingMedia.markCommitted();
    assert.ok(library.read().includes(file));
    pendingMedia.add(file, new Blob(['unsaved edit']), file);
    const preview = pendingMedia.previewUrl(file);
    library.select([file]); library.completeDelete();
    assert.deepEqual(library.read(), ['media/existing.webp']);
    assert.equal(pendingMedia.previewUrl(file), preview);
    pendingMedia.markCommitted();
    assert.deepEqual(library.read(), ['media/existing.webp', file]);
});

const pageContent = (filepath = 'content/pages/home.json') => ({
    filepath, type: 'pages', fields: { image: 'media/new.webp' },
});
const adminForPage = content => component('admin_menu.svelte', { content }, '', {
    '../../generated/media.js': { default: [] },
    '../../generated/env.js': { env: { baseurl: '/' } },
});
const visualForPage = content => component('visual_editor.svelte', {
    content, shadowContent: {}, localMediaList: [],
}, '', { '../../generated/schemas.js': { default: {} } });

await test('remounting Visual on the same page preserves deferred media', async () => {
    const content = pageContent();
    const admin = await adminForPage(content);
    const firstVisual = await visualForPage(content);
    pendingMedia.add('media/new.webp', new Blob(['new']), 'media/new.webp');
    const preview = pendingMedia.previewUrl('media/new.webp');
    firstVisual.destroy();
    const secondVisual = await visualForPage(content);
    assert.equal(pendingMedia.previewUrl('media/new.webp'), preview);
    secondVisual.destroy(); admin.destroy();
});

await test('leaving the CMS session releases its deferred previews', async () => {
    const admin = await adminForPage(pageContent());
    pendingMedia.add('media/new.webp', new Blob(['new']), 'media/new.webp');
    admin.destroy();
    assert.deepEqual(stagedPaths(), []);
});

for (const editorType of ['Code', 'Visual']) await test(editorType + ' Save supplies deferred media and marks it committed only after success', async () => {
    globalThis.FileReader = class {
        async readAsDataURL(blob) {
            this.result = 'data:' + blob.type + ';base64,' + Buffer.from(await blob.arrayBuffer()).toString('base64');
            this.onload();
        }
    };
    const content = pageContent();
    const admin = await adminForPage(content);
    pendingMedia.add('media/new.webp', new Blob(['pixels'], { type: 'image/webp' }), 'media/new.webp', 'create');
    const editor = editorType === 'Code'
        ? await component('json_editor.svelte', { content }, '')
        : await visualForPage(content);
    const save = editor.buttonProps('Save');
    const extras = await save.beforeSubmit?.() ?? [];
    assert.deepEqual(extras, [{ file: 'media/new.webp', action: 'create', encoding: 'base64', contents: 'data:image/webp;base64,cGl4ZWxz' }]);
    assert.equal(JSON.parse(save.commitList[0].contents).image, 'media/new.webp');
    let entries; pendingMedia.subscribe(value => entries = value)();
    assert.equal(entries.filter(i => !i.committed).length, 1);
    save.afterSubmit();
    assert.deepEqual(await pendingMedia.toCommitItems(), []);
    assert.ok(pendingMedia.previewUrl('media/new.webp'));
    editor.destroy(); admin.destroy();
});

await test('changing the page clears deferred media even with no editor mounted', async () => {
    const admin = await adminForPage(pageContent());
    pendingMedia.add('media/new.webp', new Blob(['new']), 'media/new.webp');
    admin.set({ content: pageContent('content/pages/other.json') });
    assert.deepEqual(stagedPaths(), []);
    admin.destroy();
});

await test('Code/Visual and View/Edit transitions keep the same page session', async () => {
    const content = pageContent();
    const admin = await adminForPage(content);
    const firstVisual = await visualForPage(content);
    pendingMedia.add('media/new.webp', new Blob(['new']), 'media/new.webp');
    const preview = pendingMedia.previewUrl('media/new.webp');
    firstVisual.destroy();
    const code = await component('json_editor.svelte', { content }, '');
    code.destroy();
    const secondVisual = await visualForPage(content);
    secondVisual.destroy();
    // View/Edit also recreates the child editor, but keeps the admin root.
    const thirdVisual = await visualForPage(content);
    assert.equal(pendingMedia.previewUrl('media/new.webp'), preview);
    thirdVisual.destroy(); admin.destroy();
});

await test('Library selection updates thumbnail under the real Svelte scheduler', async () => {
    const field = await scheduledMedia({ field: 'media/original.webp', localMediaList: [],
        changingMedia: '', uploadContext: { kind: 'standalone' }, schema: null });
    try {
        field.testPick('media/new.webp');
        await settle();
        assert.deepEqual(field.testRead(), { field: 'media/new.webp',
            fieldSrc: 'media/new.webp', displaySrc: 'media/new.webp' });
    } finally { field.$destroy(); }
});

await test('standalone save retains a committed preview before queue teardown', async () => {
    const view = await component('file_upload.svelte', {
        media: [], localMediaList: [{ file: 'media/new.webp', contents: 'data:image/webp;base64,cGl4ZWxz' }],
        mediaPrefix: '', uploadContext: { kind: 'standalone' },
    }, 'saved: addUploadsToLibrary');
    view.saved();
    assert.ok(pendingMedia.previewUrl('media/new.webp'));
    assert.deepEqual(await pendingMedia.toCommitItems(), []);
    view.destroy();
    assert.ok(pendingMedia.previewUrl('media/new.webp'));
});

await test('standalone preview preserves unrelated unsaved page media', async () => {
    pendingMedia.add('media/pending.webp', new Blob(['pending']), 'media/source.webp');
    pendingMedia.rememberSaved('media/saved.webp', 'data:image/webp;base64,cGl4ZWxz');
    let entries; pendingMedia.subscribe(value => entries = value)();
    assert.deepEqual(entries.filter(i => !i.committed).map(i => i.file), ['media/pending.webp']);
    const saved = entries.find(i => i.file === 'media/saved.webp');
    assert.equal(await saved.blob.text(), 'pixels');
    assert.equal(saved.blob.type, 'image/webp');
});

await test('standalone preview cannot replace a newer unsaved version of the same path', async () => {
    pendingMedia.add('media/same.webp', new Blob(['newer']), 'media/source.webp');
    const preview = pendingMedia.previewUrl('media/same.webp');
    pendingMedia.rememberSaved('media/same.webp', 'data:image/webp;base64,b2xkZXI=');
    assert.equal(pendingMedia.previewUrl('media/same.webp'), preview);
    let entries; pendingMedia.subscribe(value => entries = value)();
    assert.equal(entries.filter(i => !i.committed).length, 1);
});

await test('plain object media keeps metadata while its scheduled thumbnail changes', async () => {
    const field = await scheduledMedia({ field: { src: 'media/original.webp', alt: 'kept' },
        localMediaList: [], changingMedia: '', uploadContext: { kind: 'standalone' }, schema: null });
    try {
        field.testPick('media/new.webp'); await settle();
        assert.deepEqual(field.testRead(), { field: { src: 'media/new.webp', alt: 'kept' },
            fieldSrc: 'media/new.webp', displaySrc: 'media/new.webp' });
    } finally { field.$destroy(); }
});

await test('plain selection cancelled during tick cannot mutate the destroyed field', async () => {
    const field = await mediaField({ schema: null });
    field.pick('media/new.webp'); field.destroy(); await settle();
    assert.equal(field.read().field, 'media/original.webp');
});

console.log(`\n${passed} passed, ${failed} failed`);
process.exitCode = failed ? 1 : 0;
