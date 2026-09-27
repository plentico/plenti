// Dependency-free tests for the standalone upload queue state machine
// (defaults/core/cms/upload_queue.js). Run: deno run --allow-read scripts/test-upload-queue.mjs
// (also runs under node/bun).

import { readFile } from 'node:fs/promises';

let pass = 0, fail = 0;
const eq = (n, got, want) => (JSON.stringify(got) === JSON.stringify(want) ? (console.log(`  PASS  ${n}`), pass++) : (console.log(`  FAIL  ${n}\n        got ${JSON.stringify(got)}\n        want ${JSON.stringify(want)}`), fail++));
const ok = (n, v, d) => (v ? (console.log(`  PASS  ${n}`), pass++) : (console.log(`  FAIL  ${n}${d ? ' — ' + d : ''}`), fail++));

// Read the module BEFORE we stub URL (readFile needs the real URL constructor).
const src = await readFile(new URL('../defaults/core/cms/upload_queue.js', import.meta.url), 'utf8');

// URL.revokeObjectURL must exist for the module's revoke(); track calls.
// Augment the real URL (keep it a constructor) rather than replacing it.
let revoked = [];
URL.revokeObjectURL = (u) => revoked.push(u);

const { classifyFile, createUploadQueue } = await import('data:text/javascript,' + encodeURIComponent(src));

const F = (name, type) => ({ name, type: type ?? '' });
let urlSeq = 0;
const makeUrl = () => `blob:${++urlSeq}`;

console.log('=== classifyFile (MIME first, extension fallback) ===');
eq('jpeg mime -> image', classifyFile(F('x', 'image/jpeg')), 'image');
eq('png/webp/avif -> image', [classifyFile(F('a', 'image/png')), classifyFile(F('b', 'image/webp')), classifyFile(F('c', 'image/avif'))], ['image', 'image', 'image']);
eq('pdf -> passthrough', classifyFile(F('x.pdf', 'application/pdf')), 'passthrough');
eq('svg -> passthrough', classifyFile(F('x.svg', 'image/svg+xml')), 'passthrough');
eq('gif -> passthrough', classifyFile(F('x.gif', 'image/gif')), 'passthrough');
eq('pdf named .jpg (real MIME wins) -> passthrough', classifyFile(F('trick.jpg', 'application/pdf')), 'passthrough');
eq('empty MIME, .jpg ext -> image', classifyFile(F('photo.jpg', '')), 'image');
eq('octet-stream, .png ext -> image', classifyFile(F('photo.png', 'application/octet-stream')), 'image');
eq('empty MIME, .pdf ext -> passthrough', classifyFile(F('doc.pdf', '')), 'passthrough');

console.log('=== queue order + classification ===');
let q = createUploadQueue([F('a.jpg', 'image/jpeg'), F('b.pdf', 'application/pdf'), F('c.png', 'image/png')], makeUrl);
eq('deterministic order preserved', q.items.map(i => i.file.name), ['a.jpg', 'b.pdf', 'c.png']);
eq('types classified', q.items.map(i => i.type), ['image', 'passthrough', 'image']);
eq('all start queued', q.items.every(i => i.state === 'queued'), true);
ok('current is first item', q.current === q.items[0]);
ok('items carry unique ids', new Set(q.items.map(i => i.id)).size === 3);

console.log('=== canSave gating (unresolved blocks) ===');
ok('canSave false at start (nothing resolved)', q.canSave === false);
ok('resolve returns true for a member item', q.resolve(q.items[0]) === true);
ok('canSave still false (others unresolved)', q.canSave === false);
ok('unresolvedCount = 2', q.unresolvedCount === 2);
q.resolve(q.items[1]);
q.resolve(q.items[2]);
ok('canSave true when all resolved', q.canSave === true);
eq('resolvedIds in order', q.resolvedIds(), [q.items[0].id, q.items[1].id, q.items[2].id]);
ok('current null when drained', q.current === null);

console.log('=== membership guards (foreign items rejected) ===');
q = createUploadQueue([F('a.jpg', 'image/jpeg')], makeUrl);
const foreign = { id: 999999, file: F('ghost.jpg', 'image/jpeg'), type: 'image', state: 'processing', objectUrl: null, error: null };
ok('resolve(foreign) rejected', q.resolve(foreign) === false);
ok('fail(foreign) rejected', q.fail(foreign, new Error('x')) === false);
ok('prepare(foreign) rejected', q.prepare(foreign) === null);
ok('process(foreign) rejected', q.process(foreign) === null);
ok('awaitDecision(foreign) rejected', q.awaitDecision(foreign) === null);
ok('claimRun(foreign) rejected', q.claimRun(foreign) === null);
ok('foreign item state untouched', foreign.state === 'processing');
// The replaced-queue race: an OLD queue's item must be rejected by a NEW queue.
const oldQ = createUploadQueue([F('slow.pdf', 'application/pdf')], makeUrl);
const oldItem = oldQ.items[0];
oldQ.process(oldItem);
const newQ = createUploadQueue([F('fresh.jpg', 'image/jpeg')], makeUrl);
ok('NEW queue rejects the OLD queue\'s in-flight item', newQ.resolve(oldItem) === false);

console.log('=== run ownership (session-wide claims) ===');
q = createUploadQueue([F('a.jpg', 'image/jpeg'), F('b.png', 'image/png')], makeUrl);
const [a, b] = q.items;
const t1 = q.claimRun(a);
ok('claimRun returns a token', !!t1);
ok('hasActiveRun(a) true', q.hasActiveRun(a) === true);
ok('hasActiveRun(b) FALSE (claim stored with its owner)', q.hasActiveRun(b) === false);
ok('ownsRun(a, t1) true', q.ownsRun(a, t1) === true);
ok('ownsRun(b, t1) false (wrong item)', q.ownsRun(b, t1) === false);
// A LATER claim — as from a remounted component — invalidates the earlier chain.
const t2 = q.claimRun(a);
ok('later claim invalidates the earlier token', q.ownsRun(a, t1) === false && q.ownsRun(a, t2) === true);
// clearRun releases only your own claim.
q.clearRun(a, t1);                 // stale token — must be a no-op
ok('clearRun with a stale token is a no-op', q.hasActiveRun(a) === true);
q.clearRun(a, t2);
ok('clearRun with the live token releases the claim', q.hasActiveRun(a) === false);
// Terminal transitions clear the claim for their item.
const t3 = q.claimRun(a);
q.resolve(a);
ok('resolve clears the claim', q.hasActiveRun(a) === false && q.ownsRun(a, t3) === false);
const t4 = q.claimRun(b);
q.fail(b, new Error('boom'));
ok('fail clears the claim', q.ownsRun(b, t4) === false);

console.log('=== late result ignored (cancelled item) ===');
q = createUploadQueue([F('a.jpg', 'image/jpeg')], makeUrl);
const item = q.prepare(q.current); q.process(item);
const tok = q.claimRun(item);
q.cancelCurrent(item);                     // user skips mid-processing
ok('item marked cancelled', item.state === 'cancelled');
ok('cancel cleared the claim', q.ownsRun(item, tok) === false);
ok('late resolve on a cancelled item is REJECTED', q.resolve(item) === false);
ok('canSave false (nothing resolved)', q.canSave === false);

console.log('=== skipRemaining: retain resolved, drop pending AND failed ===');
revoked = [];
q = createUploadQueue(
    [F('a.jpg', 'image/jpeg'), F('b.png', 'image/png'), F('c.pdf', 'application/pdf'), F('d.gif', 'image/gif')],
    makeUrl,
);
q.resolve(q.items[0]);                              // approved
q.fail(q.items[1], new Error('read error'));        // failed
q.prepare(q.items[2]);                              // preparing (live URL)
const skipTok = q.claimRun(q.items[2]);
/* d stays queued */
q.skipRemaining();
eq('states after skipRemaining', q.items.map(i => i.state), ['resolved', 'cancelled', 'cancelled', 'cancelled']);
ok('approved item retained -> canSave TRUE', q.canSave === true);
ok('active run claim cleared', q.ownsRun(q.items[2], skipTok) === false);
ok('pending live URL revoked', revoked.length === 1);
ok('skippableCount now 0', q.skippableCount === 0);

console.log('=== failed blocks Save until removed (remove is the affordance) ===');
q = createUploadQueue([F('a.jpg', 'image/jpeg'), F('b.pdf', 'application/pdf')], makeUrl);
q.resolve(q.items[0]);
q.fail(q.items[1], new Error('boom'));
ok('canSave false while a failed item remains', q.canSave === false);
eq('failedItems() lists it', q.failedItems().map(i => i.file.name), ['b.pdf']);
q.remove(q.items[1]);
ok('canSave true after removing the failed item', q.canSave === true);
ok('failedItems() empty after remove', q.failedItems().length === 0);

console.log('=== resolved item is not revisited ===');
q = createUploadQueue([F('a.jpg', 'image/jpeg'), F('b.png', 'image/png')], makeUrl);
q.resolve(q.items[0]);
ok('current skips the resolved item -> points at b', q.current === q.items[1]);

console.log('=== object-URL lifecycle: destroy with LIVE urls, called twice ===');
revoked = [];
q = createUploadQueue([F('a.jpg', 'image/jpeg'), F('b.png', 'image/png')], makeUrl);
q.prepare(q.items[0]);
q.prepare(q.items[1]);                     // two LIVE urls outstanding
const dTok = q.claimRun(q.items[0]);
q.destroy();
ok('destroy revoked BOTH live urls', revoked.length === 2);
ok('destroy cleared the refs', q.items.every(i => i.objectUrl === null));
ok('destroy cleared the run claim', q.ownsRun(q.items[0], dTok) === false);
q.destroy();                               // second call — idempotent
ok('second destroy revokes nothing further', revoked.length === 2);

console.log('=== resolve revokes + revocation is ref-clearing ===');
revoked = [];
q = createUploadQueue([F('a.jpg', 'image/jpeg')], makeUrl);
const it = q.prepare(q.current);
ok('prepare set objectUrl', !!it.objectUrl && it.state === 'preparing');
q.resolve(it);
ok('resolve revoked the url and cleared the ref', revoked.length === 1 && it.objectUrl === null);
q.destroy();
ok('destroy after resolve re-revokes nothing (ref already null)', revoked.length === 1);

console.log('=== skippableCount drives the affordance ===');
q = createUploadQueue([F('a.jpg', 'image/jpeg'), F('b.pdf', 'application/pdf'), F('c.png', 'image/png')], makeUrl);
ok('3 skippable at start', q.skippableCount === 3);
q.resolve(q.items[1]);
ok('2 skippable after one resolve', q.skippableCount === 2);
q.fail(q.items[2], new Error('x'));
ok('failed still counts as skippable', q.skippableCount === 2);
q.cancelCurrent(q.items[0]);
ok('cancelled does not count', q.skippableCount === 1);

console.log(`\n${pass} passed, ${fail} failed`);
if (typeof process !== 'undefined') process.exit(fail ? 1 : 0);
else if (typeof Deno !== 'undefined') Deno.exit(fail ? 1 : 0);
