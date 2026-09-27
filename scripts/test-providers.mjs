// Dependency-free contract tests for the CMS commit providers
// (defaults/core/cms/providers/{gitlab,gitea,local}.js).
//
//   Run:  node scripts/test-providers.mjs   (also runs under: deno run -A, bun)
//
// No test framework, no npm install — same approach as test-crop-engine.mjs.
// The providers import `generated/env.js`, `svelte/store`, and sibling CMS modules
// that only exist inside a Plenti build, so we load each provider's *source*,
// rewrite those imports to inline stubs, and import the result through a data: URL
// (which always parses as a module regardless of repo config).
//
// `fetch`, `history`, and `location` are mocked just enough to drive — and assert —
// the request contract each provider emits. Focus: the provider-neutral `upsert`
// resolution (create-vs-update) and Gitea's media-before-content ordering added to
// close the remote gaps for #364 (see ADR 0001).

import { readFile } from 'node:fs/promises';

let pass = 0, fail = 0;
const ok = (n) => { console.log(`  PASS  ${n}`); pass++; };
const bad = (n, d) => { console.log(`  FAIL  ${n}${d ? `\n        ${d}` : ''}`); fail++; };
const eq = (n, got, want) => (JSON.stringify(got) === JSON.stringify(want) ? ok(n) : bad(n, `got ${JSON.stringify(got)} want ${JSON.stringify(want)}`));
const truthy = (n, v, d) => (v ? ok(n) : bad(n, d));

// --- minimal DOM/history so the providers' post-success block doesn't throw ---
globalThis.history = { pushState() {} };
globalThis.location = { pathname: '/' };

// --- load a provider's source with its build-only imports stubbed ---
async function loadProvider(relPath, { branch = 'main', repo = 'https://git.example.com/owner/repo' } = {}) {
    let src = await readFile(new URL(`../defaults/core/cms/providers/${relPath}`, import.meta.url), 'utf8');
    // Replace every build-only import line with inline stubs. Providers vary in what
    // they import from ../url_checker.js (gitea/gitlab take makeUrl+normalizeRoute;
    // local takes only normalizeRoute), so stub any `{ ... } from '../url_checker.js'`
    // line generically, and prepend the helpers both forms may reference.
    src = src
        .replace(/import \{ env \} from '[^']*';/, `const env = { cms: { repo: ${JSON.stringify(repo)}, branch: ${JSON.stringify(branch)} }, baseurl: '', local: false };`)
        .replace(/import \{[^}]*\} from '[^']*url_checker\.js';/, 'const makeUrl = (u) => new URL(u); const normalizeRoute = (r) => r;')
        .replace(/import evaluateRoute from '[^']*';/, 'const evaluateRoute = () => "/x";');
    return import('data:text/javascript,' + encodeURIComponent(src));
}

// --- a fake `user` store (Svelte-store shape: subscribe(fn) -> unsub) ---
const fakeUser = {
    subscribe(fn) { fn({ isAuthenticated: true, tokens: { access_token: 'tok' } }); return () => {}; },
};

// --- a shadowContent spy that counts the commit-level success/delete callbacks,
//     so we can assert onSave fires once per *commit*, never per file ---
const spyShadow = () => {
    const counts = { saves: 0, deletes: 0 };
    return { sc: { onSave: () => { counts.saves++; }, onDelete: () => { counts.deletes++; } }, counts };
};

// --- a scripted fetch: each call consumes the next handler from a queue, and
//     every request is recorded for assertions ---
function scriptFetch(handlers) {
    const calls = [];
    let i = 0;
    globalThis.fetch = async (url, opts = {}) => {
        const h = handlers[i++] ?? (() => { throw new Error(`unexpected fetch #${i}: ${opts.method} ${url}`); });
        calls.push({ url: String(url), method: opts.method, body: opts.body ? JSON.parse(opts.body) : undefined });
        return h(String(url), opts);
    };
    return calls;
}
const res = (status, { json, headers } = {}) => ({
    ok: status >= 200 && status < 300,
    status,
    json: async () => json ?? {},
    text: async () => '',
    headers: new Map(Object.entries(headers ?? {})),
});

// ─────────────────────────── GitLab ───────────────────────────
console.log('=== GitLab (atomic batch; upsert via HEAD) ===');
{
    const { commitGitlab } = await loadProvider('gitlab.js');

    // absent media → create; the /user HEAD-resolve then the single commit POST
    let calls = scriptFetch([
        () => res(404),                                   // HEAD media (absent)
        () => res(201, { json: {} }),                     // POST commit
    ]);
    await commitGitlab(
        [{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' },
         { file: 'page.json', action: 'update', encoding: 'text', contents: '{}' }],
        null, 'update', 'text', fakeUser,
    );
    const head = calls.find(c => c.method === 'HEAD');
    const commit = calls.find(c => c.method === 'POST');
    truthy('absent media HEADed then committed', head && commit);
    const aAction = commit.body.actions.find(a => a.file_path === 'media/a.webp').action;
    eq('absent derivative → action create', aAction, 'create');
    eq('content + media in ONE commit request (one POST)', calls.filter(c => c.method === 'POST').length, 1);
    eq('one commit carries BOTH files', commit.body.actions.length, 2);

    // existing media → update + last_commit_id
    calls = scriptFetch([
        () => res(200, { headers: { 'X-Gitlab-Last-Commit-Id': 'abc123' } }), // HEAD (exists)
        () => res(201, { json: {} }),                                          // POST commit
    ]);
    await commitGitlab(
        [{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' }],
        null, 'update', 'text', fakeUser,
    );
    const act = calls.find(c => c.method === 'POST').body.actions[0];
    eq('existing derivative → action update', act.action, 'update');
    eq('update carries last_commit_id', act.last_commit_id, 'abc123');

    // metadata error other than 404 → abort, no commit
    calls = scriptFetch([() => res(403)]); // HEAD 403
    let threw = false;
    try { await commitGitlab([{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:,AAAA' }], null, 'update', 'text', fakeUser); }
    catch { threw = true; }
    truthy('HEAD 403 aborts (not treated as absent)', threw);
    eq('no commit POST after metadata abort', calls.filter(c => c.method === 'POST').length, 0);
}

// ─────────────────────────── Gitea ───────────────────────────
console.log('=== Gitea (sequential per-file; upsert via GET; media before content) ===');
{
    const { commitGitea } = await loadProvider('gitea.js');

    // ordering + create: media (absent) committed before content
    let calls = scriptFetch([
        () => res(200, { json: { login: 'u', email: 'u@x' } }),  // GET /user
        () => res(404),                                          // GET media (absent → create)
        () => res(201, { json: {} }),                            // POST media
        () => res(200, { json: { sha: 'c0ffee' } }),             // GET content (update sha)
        () => res(200, { json: {} }),                            // PUT content
    ]);
    await commitGitea(
        [{ file: 'page.json', action: 'update', encoding: 'text', contents: '{}' },
         { file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' }],
        null, 'update', 'text', fakeUser,
    );
    const writes = calls.filter(c => c.method === 'POST' || c.method === 'PUT' || c.method === 'DELETE');
    truthy('media write occurs before content write',
        writes[0].url.includes('media/a.webp') && writes[1].url.includes('page.json'),
        `order was ${writes.map(w => w.url.split('/contents/')[1]).join(', ')}`);
    eq('absent derivative → POST (create)', writes[0].method, 'POST');

    // existing media → PUT with sha
    calls = scriptFetch([
        () => res(200, { json: { login: 'u' } }),          // GET /user
        () => res(200, { json: { sha: 'deadbeef' } }),     // GET media (exists → update)
        () => res(200, { json: {} }),                      // PUT media
    ]);
    await commitGitea(
        [{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' }],
        null, 'update', 'text', fakeUser,
    );
    const mediaWrite = calls.find(c => c.method === 'PUT');
    truthy('existing derivative → PUT (update)', mediaWrite);
    eq('PUT carries the resolved sha', mediaWrite.body.sha, 'deadbeef');

    // media failure → content never attempted
    calls = scriptFetch([
        () => res(200, { json: { login: 'u' } }),          // GET /user
        () => res(404),                                    // GET media (create)
        () => res(422, { json: { message: 'boom' } }),     // POST media FAILS
        // (no further handlers — a content write would throw "unexpected fetch")
    ]);
    let threw = false;
    try {
        await commitGitea(
            [{ file: 'page.json', action: 'update', encoding: 'text', contents: '{}' },
             { file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:,AAAA' }],
            null, 'update', 'text', fakeUser,
        );
    } catch { threw = true; }
    truthy('media failure throws', threw);
    truthy('content write NEVER attempted after media failure',
        !calls.some(c => (c.method === 'PUT' || c.method === 'POST') && c.url.includes('page.json')));

    // A deferred as-is upload uses CREATE, not upsert. It must fail before
    // content is written too. Route-aware responses let either ordering run,
    // so a failure here is the dangerous content write, not a mock mismatch.
    calls = scriptFetch(Array(8).fill((url, opts) => {
        if (url.endsWith('/user')) return res(200, { json: { login: 'u' } });
        if (opts.method === 'GET') return res(200, { json: { sha: 'content-sha' } });
        if (url.endsWith('/media/new.webp')) return res(500, { json: { message: 'raw write failed' } });
        return res(200, { json: {} });
    }));
    let rawError;
    try {
        await commitGitea([
            { file: 'content/pages/test.json', contents: '{"image":"media/new.webp"}' },
            { file: 'media/new.webp', action: 'create', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' },
        ], null, 'update', 'text', fakeUser);
    } catch (error) { rawError = error.message; }
    eq('raw create failure surfaces its provider error', rawError, 'Publish failed: raw write failed');
    eq('content write NEVER attempted after raw create failure',
        calls.filter(c => ['PUT', 'POST'].includes(c.method) && c.url.includes('/content/pages/test.json')).length, 0);

    // Stable mixed-media ordering, preserving create/update semantics.
    calls = scriptFetch(Array(12).fill((url, opts) => {
        if (url.endsWith('/user')) return res(200, { json: { login: 'u' } });
        if (opts.method === 'GET') return res(200, { json: { sha: 'existing-sha' } });
        return res(200, { json: {} });
    }));
    await commitGitea([
        { file: 'content/pages/test.json', contents: '{}' },
        { file: 'media/raw.pdf', action: 'create', encoding: 'base64', contents: 'data:application/pdf;base64,AAAA' },
        { file: 'media/derivative.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' },
        { file: '/media/nested/as-is.webp', action: 'create', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' },
    ], null, 'update', 'text', fakeUser);
    const mixedWrites = calls.filter(c => ['POST', 'PUT'].includes(c.method));
    eq('all mixed media writes precede content in original media order',
        mixedWrites.map(c => c.url.split('/contents/')[1]),
        ['media/raw.pdf', 'media/derivative.webp', '/media/nested/as-is.webp', 'content/pages/test.json']);
    eq('ordering preserves raw create and resolved derivative update methods',
        mixedWrites.map(c => c.method), ['POST', 'PUT', 'POST', 'PUT']);

    // Creating a new page must also wait for a passthrough PDF to persist.
    calls = scriptFetch(Array(8).fill((url, opts) => {
        if (url.endsWith('/user')) return res(200, { json: { login: 'u' } });
        if (url.endsWith('/media/raw.pdf')) return res(500, { json: { message: 'PDF write failed' } });
        if (opts.method === 'GET') return res(404);
        return res(201, { json: {} });
    }));
    const rawFailure = spyShadow();
    try {
        await commitGitea([
            { file: 'content/pages/new.json', contents: '{"document":"media/raw.pdf"}' },
            { file: 'media/raw.pdf', action: 'create', encoding: 'base64', contents: 'data:application/pdf;base64,AAAA' },
        ], rawFailure.sc, 'create', 'text', fakeUser);
    } catch { /* expected */ }
    eq('new content create NEVER attempted after passthrough failure',
        calls.filter(c => c.method === 'POST' && c.url.includes('/content/pages/new.json')).length, 0);
    eq('passthrough failure never signals save success', rawFailure.counts.saves, 0);

    // resolve error (non-404) on upsert → abort
    calls = scriptFetch([
        () => res(200, { json: { login: 'u' } }),          // GET /user
        () => res(401),                                    // GET media 401 → abort
    ]);
    threw = false;
    try { await commitGitea([{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:,AAAA' }], null, 'update', 'text', fakeUser); }
    catch { threw = true; }
    truthy('GET 401 on upsert aborts (not treated as absent)', threw);

    // backward-compat: an explicit content-only update still works as before
    calls = scriptFetch([
        () => res(200, { json: { login: 'u' } }),          // GET /user
        () => res(200, { json: { sha: 'aaa' } }),          // GET content sha
        () => res(200, { json: {} }),                      // PUT content
    ]);
    await commitGitea([{ file: 'page.json', encoding: 'text', contents: '{}' }], null, 'update', 'text', fakeUser);
    eq('explicit update unchanged (PUT with sha)', calls.find(c => c.method === 'PUT')?.body.sha, 'aaa');

    // onSave is a COMMIT-level signal: it must fire once after every file commits,
    // never per file — otherwise media-first ordering would signal "saved" before a
    // failing content write, breaking the retryable guarantee.
    // (1) multi-item success → onSave exactly once
    {
        const { sc, counts } = spyShadow();
        scriptFetch([
            () => res(200, { json: { login: 'u' } }),       // GET /user
            () => res(404),                                 // GET media (absent → create)
            () => res(201, { json: {} }),                   // POST media
            () => res(200, { json: { sha: 'c0ffee' } }),    // GET content sha
            () => res(200, { json: {} }),                   // PUT content
        ]);
        await commitGitea(
            [{ file: 'page.json', action: 'update', encoding: 'text', contents: '{}' },
             { file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' }],
            sc, 'update', 'text', fakeUser,
        );
        eq('multi-item success → onSave called exactly once', counts.saves, 1);
    }
    // (2) media failure → onSave zero times
    {
        const { sc, counts } = spyShadow();
        scriptFetch([
            () => res(200, { json: { login: 'u' } }),       // GET /user
            () => res(404),                                 // GET media (create)
            () => res(422, { json: { message: 'boom' } }),  // POST media FAILS
        ]);
        try {
            await commitGitea(
                [{ file: 'page.json', action: 'update', encoding: 'text', contents: '{}' },
                 { file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:,AAAA' }],
                sc, 'update', 'text', fakeUser,
            );
        } catch { /* expected */ }
        eq('media failure → onSave never called', counts.saves, 0);
    }
    // (3) content failure AFTER successful media → onSave zero times (retryable)
    {
        const { sc, counts } = spyShadow();
        scriptFetch([
            () => res(200, { json: { login: 'u' } }),       // GET /user
            () => res(404),                                 // GET media (create)
            () => res(201, { json: {} }),                   // POST media OK
            () => res(200, { json: { sha: 'c0ffee' } }),    // GET content sha
            () => res(500, { json: { message: 'server' } }),// PUT content FAILS
        ]);
        try {
            await commitGitea(
                [{ file: 'page.json', action: 'update', encoding: 'text', contents: '{}' },
                 { file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:,AAAA' }],
                sc, 'update', 'text', fakeUser,
            );
        } catch { /* expected */ }
        eq('content failure after successful media → onSave never called', counts.saves, 0);
    }
    // (4) single-item content save unchanged → onSave exactly once
    {
        const { sc, counts } = spyShadow();
        scriptFetch([
            () => res(200, { json: { login: 'u' } }),       // GET /user
            () => res(200, { json: { sha: 'aaa' } }),       // GET content sha
            () => res(200, { json: {} }),                   // PUT content
        ]);
        await commitGitea([{ file: 'page.json', action: 'update', encoding: 'text', contents: '{}' }], sc, 'update', 'text', fakeUser);
        eq('single-item save → onSave called exactly once', counts.saves, 1);
    }
}

// ─────────────────────────── Local ───────────────────────────
console.log('=== Local (/postlocal; upsert maps to create) ===');
{
    const { postLocal } = await loadProvider('local.js');
    let calls = scriptFetch([() => res(200)]); // /postlocal
    await postLocal(
        [{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:image/webp;base64,AAAA' }],
        null, 'update', 'text',
    );
    const body = calls[0].body;
    eq('upsert sent to /postlocal as create', body[0].action, 'create');

    // a non-2xx local response surfaces as a thrown (failed) save via .text()
    calls = scriptFetch([() => ({ ok: false, status: 400, text: async () => 'bad path' })]);
    let threw = false;
    try { await postLocal([{ file: 'media/a.webp', action: 'upsert', encoding: 'base64', contents: 'data:,AAAA' }], null, 'update', 'text'); }
    catch (e) { threw = /bad path/.test(e.message); }
    truthy('local non-2xx → thrown failed save (plain-text message)', threw);
}

console.log(`\n=== ${pass} passed, ${fail} failed ===`);
if (typeof process !== 'undefined') process.exit(fail ? 1 : 0);
else if (typeof Deno !== 'undefined') Deno.exit(fail ? 1 : 0);
