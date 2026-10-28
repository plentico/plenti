// Run: node scripts/test-cms-nested-arrays.mjs
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

class Element {
    constructor(tag) {
        this.tag = tag;
        this.children = [];
        this.className = '';
        this.style = {};
        this.attributes = {};
        this.textContent = '';
    }
    appendChild(child) { this.children.push(child); return child; }
    addEventListener() {}
    setAttribute(name, value) { this.attributes[name] = value; }
    set innerHTML(value) { if (value === '') this.children = []; }
}
const text = node => node.text ? node.text : [node.textContent, ...node.children.map(text)].join('');
const descendants = (node, predicate, found = []) => {
    if (predicate(node)) found.push(node);
    (node.children || []).forEach(child => descendants(child, predicate, found));
    return found;
};
const byClass = (node, name) => descendants(node, child => child.className?.split(/\s+/).includes(name));

const document = {
    readyState: 'loading',
    addEventListener() {},
    createElement: tag => new Element(tag),
    createTextNode: value => ({ text: value }),
};
const window = {};
let source = await readFile(new URL('../defaults/core/cms/cms.js', import.meta.url), 'utf8');
source = source.replace(/\}\)\(\);\s*$/, '  window.arrayTest = { renderArrayField };\n})();');
vm.runInNewContext(source, { window, document });

const rows = [[
    { image: { alt: '', height: '768', url: '/image.jpg', width: '1024' }, title: 'Brain & Body', url: '/brain-body' },
    { image: { alt: 'Alt', height: '2400', url: '/image-2.jpg', width: '3600' }, title: 'Public Health', url: '/public-health' },
    { image: null, title: 'Notes', url: '/notes' },
]];
const tray = new Element('div');
const root = window.arrayTest.renderArrayField(tray, null, 'consequence_rows', 'consequence_rows', rows, () => {}, {});

assert.equal(byClass(root, 'plenti-array').length, 1, 'outer row group has no redundant nested Items wrapper');
assert.equal(byClass(root, 'plenti-array-item').length, 4, 'one row plus three reorderable cards');
assert.equal(byClass(root, 'plenti-array-items').every(group => group.style.display !== 'none'), true, 'outer cards exclusively control nested content visibility');
assert.equal(byClass(root, 'plenti-array-remove').length, 4, 'nested cards have remove controls');
const labels = byClass(root, 'plenti-field-nested').map(field => text(field.children[0]));
assert.deepEqual(labels, ['image', 'image'], 'object values render labeled nested fieldsets');
const scalarLabels = descendants(root, node => node.tag === 'label').map(text);
for (const name of ['alt', 'height', 'url', 'width', 'title']) {
    assert.ok(scalarLabels.includes(name), name + ' field is rendered');
}
assert.equal(byClass(root, 'plenti-field').filter(field => text(field.children[0]) === 'image' && !field.className.includes('plenti-field-nested')).length, 1, 'null image remains editable');
console.log('PASS nested arrays render complete reorderable/removable component groups');
console.log('PASS nested objects render recursively, including null values');
