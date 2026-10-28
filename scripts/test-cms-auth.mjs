// Run: node scripts/test-cms-auth.mjs
// Verifies that embedded CMS DOM state follows authentication, including when
// Pattr rewrites body.className while updating a reactive user binding.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

class ClassList {
    constructor() { this.values = new Set(); }
    add(...names) { names.forEach(name => this.values.add(name)); }
    remove(...names) { names.forEach(name => this.values.delete(name)); }
    toggle(name, force) {
        const enabled = force === undefined ? !this.contains(name) : force;
        if (enabled) this.add(name); else this.remove(name);
        return enabled;
    }
    contains(name) { return this.values.has(name); }
}

class Element {
    constructor(tag) {
        this.tag = tag;
        this.children = [];
        this.classList = new ClassList();
        this.style = {};
        this.attributes = {};
        this.textContent = '';
    }
    set className(value) {
        this.classList.values = new Set(value.split(/\s+/).filter(Boolean));
    }
    get className() { return [...this.classList.values].join(' '); }
    set innerHTML(value) { if (value === '') this.children = []; }
    appendChild(child) { this.children.push(child); return child; }
    addEventListener() {}
    setAttribute(name, value) { this.attributes[name] = value; }
    querySelector(selector) {
        if (selector === '.plenti-admin-bar') {
            return this.children.find(child => child.classList.contains('plenti-admin-bar')) || null;
        }
        return null;
    }
}

async function scenario(reactiveBody) {
    const body = new Element('body');
    body.className = 'site-class';
    let container = new Element('div');
    const storage = new Map([['plenti_cms_local_logout', '1']]);
    const localStorage = {
        getItem: key => storage.get(key) || null,
        setItem: (key, value) => storage.set(key, String(value)),
        removeItem: key => storage.delete(key),
    };
    const data = { user: { loggedIn: false }, content: { fields: { title: 'Original' } } };
    const window = {
        Pattr: { data: {} },
        location: { pathname: '/' },
        dispatchEvent() {},
    };
    Object.defineProperty(window.Pattr, 'data', {
        get: () => data,
        set: value => Object.assign(data, value),
    });
    // Model the two relevant Pattr effects: conditional CMS containers and a
    // body p-class binding that replaces classes managed by other scripts.
    const originalUserDescriptor = Object.getOwnPropertyDescriptor(data, 'user');
    Object.defineProperty(data, 'user', {
        configurable: true,
        get: originalUserDescriptor.get || (() => originalUserDescriptor.value),
        set: user => {
            if (originalUserDescriptor.set) originalUserDescriptor.set(user);
            else originalUserDescriptor.value = user;
            if (reactiveBody) {
                container = null;
                body.className = 'site-class';
                // Pattr can render conditional content after the data setter.
                setTimeout(() => {
                    container = user.loggedIn ? new Element('div') : null;
                }, 0);
            }
        },
    });
    const document = {
        readyState: 'complete',
        body,
        createElement: tag => new Element(tag),
        createTextNode: text => ({ text }),
        querySelector: selector => selector === '[plenti-cms]' ? container : null,
    };
    const context = {
        window, document, localStorage,
        sessionStorage: localStorage,
        CustomEvent: class {},
        fetch: async url => {
            assert.equal(url, '/admin/data.json');
            return { json: async () => ({ env: { local: true, cms: { provider: 'local' } }, content: [] }) };
        },
        console, URL, setTimeout, clearTimeout,
    };
    let source = await readFile(new URL('../defaults/core/cms/cms.js', import.meta.url), 'utf8');
    source = source.replace(/\}\)\(\);\s*$/, '  window.cmsTest = { liveUpdate, setTrayOpen, setTrayMode, state, syncBodyClasses };\n})();');
    vm.runInNewContext(source, context);
    await new Promise(resolve => setImmediate(resolve));

    assert.equal(body.classList.contains('plenti-admin-active'), false, 'logged out at boot');
    window.user.login();
    await new Promise(resolve => setTimeout(resolve, 10));
    assert.equal(body.classList.contains('plenti-admin-active'), true, 'logged in');
    assert.ok(container.querySelector('.plenti-admin-bar'), 'admin bar mounted');
    window.user.logout();
    assert.equal(body.classList.contains('plenti-admin-active'), false, 'logged out');
    window.user.login();
    await new Promise(resolve => setTimeout(resolve, 10));
    assert.equal(body.classList.contains('plenti-admin-active'), true, 'logged in again');

    window.cmsTest.setTrayOpen(true);
    assert.ok(body.classList.contains('plenti-tray-open'));
    assert.ok(body.classList.contains('plenti-tray-form'));
    window.cmsTest.state.trayInfoMode = true;
    window.cmsTest.syncBodyClasses();
    assert.ok(body.classList.contains('plenti-tray-info'));
    assert.equal(body.classList.contains('plenti-tray-form'), false);
    window.cmsTest.state.trayInfoMode = false;
    window.cmsTest.state.trayJsonMode = true;
    window.cmsTest.syncBodyClasses();
    assert.ok(body.classList.contains('plenti-tray-code'));
    assert.equal(body.classList.contains('plenti-tray-info'), false);

    // Model Pattr replacing body classes when a JSON edit reassigns content.
    body.className = 'site-class';
    window.cmsTest.liveUpdate('title', 'Edited');
    await new Promise(resolve => setTimeout(resolve, 10));
    assert.ok(body.classList.contains('plenti-admin-active'), 'live update preserves admin state');
    assert.ok(body.classList.contains('plenti-tray-open'), 'live update preserves tray state');
    assert.ok(body.classList.contains('plenti-tray-code'), 'live update preserves JSON view state');
    assert.ok(body.classList.contains('site-class'), 'live update preserves site classes');

    window.cmsTest.setTrayOpen(false);
    for (const name of ['plenti-tray-open', 'plenti-tray-form', 'plenti-tray-info', 'plenti-tray-code']) {
        assert.equal(body.classList.contains(name), false, name + ' removed when tray closes');
    }
}

await scenario(true);
console.log('PASS auth state survives reactive body class updates');
await scenario(false);
console.log('PASS auth state with a static body and persistent CMS container');
console.log('PASS tray view classes and JSON live updates preserve body state');
