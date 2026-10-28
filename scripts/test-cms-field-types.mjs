// Run: node scripts/test-cms-field-types.mjs
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

let source = await readFile(new URL('../defaults/core/cms/cms.js', import.meta.url), 'utf8');
source = source.replace(/\}\)\(\);\s*$/, '  window.fieldTypes = { schemaType, isMediaPath, isStructuredArray };\n})();');
const window = {};
vm.runInNewContext(source, {
    window,
    document: { readyState: 'loading', addEventListener() {} },
});
const { schemaType, isMediaPath, isStructuredArray } = window.fieldTypes;

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
