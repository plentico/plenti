// Run: node scripts/test-cms-field-types.mjs
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

let source = await readFile(new URL('../defaults/core/cms/cms.js', import.meta.url), 'utf8');
source = source.replace(/\}\)\(\);\s*$/, '  window.fieldTypes = { schemaType, isMediaPath };\n})();');
const window = {};
vm.runInNewContext(source, {
    window,
    document: { readyState: 'loading', addEventListener() {} },
});
const { schemaType, isMediaPath } = window.fieldTypes;

assert.equal(schemaType({}, 'count', 165), 'number');
assert.equal(schemaType({}, 'settings', { enabled: true }), 'object');
assert.equal(schemaType({}, 'hero', '/media/hero.WEBP'), 'media');
assert.equal(schemaType({}, 'hero', 'media/document.pdf'), 'media');
assert.equal(schemaType({}, 'hero', 'media/file.txt'), 'text');
assert.equal(schemaType({ hero: 'text' }, 'hero', '/media/hero.png'), 'text');
assert.equal(schemaType({ 'settings.count': { type: 'text' } }, 'count', 165, 'settings.count'), 'text');
assert.equal(isMediaPath('/assets/image.png'), false);
console.log('PASS automatic number, object, and media field type inference');
console.log('PASS explicit and nested schema overrides take precedence');
