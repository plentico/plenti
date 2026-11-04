// Run: node scripts/test-cms-field-types.mjs
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

let source = await readFile(new URL('../defaults/core/cms/cms.js', import.meta.url), 'utf8');
source = source.replace(/\}\)\(\);\s*$/, '  window.fieldTypes = { schemaType, isMediaPath, isStructuredArray, fieldOptions, fieldChoices, referenceValue, referenceItems, applyFieldDefaults, setData: data => { state.data = data; } };\n})();');
const window = {};
vm.runInNewContext(source, {
    window,
    document: { readyState: 'loading', addEventListener() {} },
});
const { schemaType, isMediaPath, isStructuredArray, fieldOptions, fieldChoices, referenceValue, referenceItems, applyFieldDefaults } = window.fieldTypes;

assert.equal(schemaType({}, 'count', 165), 'number');
assert.equal(schemaType({}, 'settings', { enabled: true }), 'object');
assert.equal(schemaType({}, 'hero', '/media/hero.WEBP'), 'media');
assert.equal(schemaType({}, 'hero', 'media/document.pdf'), 'media');
assert.equal(schemaType({}, 'hero', 'media/file.txt'), 'text');
assert.equal(schemaType({ hero: 'text' }, 'hero', '/media/hero.png'), 'text');
assert.equal(schemaType({ 'settings.count': { type: 'text' } }, 'count', 165, 'settings.count'), 'text');
assert.equal(isMediaPath('/assets/image.png'), false);
const consequenceRows = [[
    { image: { alt: '', url: '/media/card.jpg' }, title: 'Brain & Body Alterations' },
    { image: { alt: '', url: '/media/card-2.jpg' }, title: 'Public Health' },
]];
assert.equal(isStructuredArray(consequenceRows), true, 'arrays of arrays use recursive groups');
assert.equal(isStructuredArray(consequenceRows[0]), true, 'arrays of objects use recursive groups');
assert.equal(isStructuredArray([{ title: 'Card' }, 'not an object']), false, 'mixed arrays retain scalar behavior');
console.log('PASS automatic number, object, and media field type inference');
console.log('PASS explicit and nested schema overrides take precedence');
console.log('PASS nested arrays are treated as recursive structured groups');

const schema = {
    status: { type: 'select', options: { choices: ['draft', 'published'], default: 'draft' } },
    featured: { type: 'boolean', options: { disabled: true, hidden: true, required: true } },
};
assert.deepEqual(fieldOptions(schema, 'featured'), { disabled: true, hidden: true, required: true });
assert.deepEqual(Array.from(fieldChoices(schema, 'status')), ['draft', 'published']);
assert.deepEqual(fieldChoices({ status: { options: ['draft', 'published'] } }, 'status'), ['draft', 'published'], 'legacy choice arrays remain compatible');
assert.deepEqual(fieldOptions({ image: { options: [{ width: 500 }] } }, 'image'), { width: 500 }, 'legacy media option object remains compatible');
const blog = { type: 'blog', path: '/blog/post', fields: { title: 'A post' } };
const page = { type: 'pages', path: '/about', fields: { title: 'About' } };
window.fieldTypes.setData({ content: [blog, page], schemas: { pages: schema }, defaults: { pages: { title: 'Existing' } } });
assert.deepEqual(Array.from(referenceItems({ related: { options: { types: ['blog'], search: ['fields.title'] } } }, 'related', null, 'post')), [blog]);
assert.equal(referenceValue(blog, 'fields.title'), 'A post');
assert.equal(referenceValue(blog, 'path'), '/blog/post');
const defaults = applyFieldDefaults('pages', {});
assert.equal(defaults.status, 'draft');
assert.equal(applyFieldDefaults('pages', { status: 'published' }).status, 'published', 'explicit defaults win');
console.log('PASS object options, legacy migration, references, and draft defaults');
