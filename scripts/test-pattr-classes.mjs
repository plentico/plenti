// Run: node scripts/test-pattr-classes.mjs
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

const window = {};
const document = {
    documentElement: { tagName: 'HTML', attributes: [], children: [], hasAttribute: () => false },
    getElementById: () => null,
};
const source = await readFile(new URL('../defaults/core/pattr.js', import.meta.url), 'utf8');
vm.runInNewContext(source.replace(',window.Pattr.start();', ''), {
    window, document, console, setTimeout, clearTimeout,
});
const setClass = window.Pattr.directives['p-class'];
const body = { className: 'site-class plenti-admin-active plenti-tray-open plenti-tray-form' };

setClass(body, { 'user-logged-in': true, 'menu-open': false });
assert.equal(body.className, 'site-class plenti-admin-active plenti-tray-open plenti-tray-form user-logged-in');
for (let i = 0; i < 20; i++) {
    // Each call models the body binding being walked for another keystroke.
    setClass(body, { 'user-logged-in': true, 'menu-open': false });
    assert.ok(body.className.includes('plenti-admin-active'));
    assert.ok(body.className.includes('plenti-tray-open'));
    assert.ok(body.className.includes('plenti-tray-form'));
}
body.className += ' plenti-tray-code';
setClass(body, { 'user-logged-in': false, 'menu-open': true });
assert.equal(body.className, 'site-class plenti-admin-active plenti-tray-open plenti-tray-form menu-open plenti-tray-code');
assert.ok(body.className.includes('menu-open'), 'new reactive classes are applied');
assert.equal(body.className.includes('user-logged-in'), false, 'old reactive classes are removed');
assert.ok(body.className.includes('site-class'), 'original static classes survive');

const replacing = { className: 'initial' };
setClass(replacing, ['one', 'two'], { replace: [] });
assert.equal(replacing.className, 'one two', 'replace modifier remains authoritative');
console.log('PASS p-class preserves external CMS classes across repeated renders');
console.log('PASS p-class still updates reactive classes and honors replace');
