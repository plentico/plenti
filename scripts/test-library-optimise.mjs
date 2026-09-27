// Dependency-free tests for the Media Library gateway helpers
// (defaults/core/cms/library_optimise.js). Run: node scripts/test-library-optimise.mjs
// (also deno run -A). Uses Web Crypto (available in node >=15, deno, browsers).

import { readFile } from 'node:fs/promises';

let pass = 0, fail = 0;
const eq = (n, got, want) => (JSON.stringify(got) === JSON.stringify(want) ? (console.log(`  PASS  ${n}`), pass++) : (console.log(`  FAIL  ${n}\n        got ${JSON.stringify(got)}\n        want ${JSON.stringify(want)}`), fail++));
const ok = (n, v, d) => (v ? (console.log(`  PASS  ${n}`), pass++) : (console.log(`  FAIL  ${n}${d ? ' — ' + d : ''}`), fail++));

// load the module source (no build-only imports here, so a plain data: import works)
const src = await readFile(new URL('../defaults/core/cms/library_optimise.js', import.meta.url), 'utf8');
const { libraryFingerprint, libraryOutputPath, extForMime } = await import('data:text/javascript,' + encodeURIComponent(src));

// a fake File: arrayBuffer() over given bytes
const fakeFile = (bytes) => ({ arrayBuffer: async () => new Uint8Array(bytes).buffer });

const OPTS = { crop: false, scale: true, convert: 'webp', quality: 0.82, maxWidth: 2048, maxHeight: 2048 };
const RECT = { x: 0, y: 0, width: 5000, height: 3000 };

console.log('=== extForMime ===');
eq('webp', extForMime('image/webp'), 'webp');
eq('png fallback', extForMime('image/png'), 'png');
eq('unknown -> png', extForMime('image/gif'), 'png');

console.log('=== libraryOutputPath shape ===');
const p = libraryOutputPath({ sourcePath: 'media/photo.jpg', fingerprint: 'a81c92f4deadbeef0000', width: 2048, height: 1229, mime: 'image/webp' });
eq('hashed name shape', p, 'media/photo-a81c92f4deadbeef-2048x1229.webp');
// PNG fallback names .png even if requested webp
eq('fallback ext from mime', libraryOutputPath({ sourcePath: 'media/photo.jpg', fingerprint: 'deadbeefdeadbeef', width: 100, height: 50, mime: 'image/png' }), 'media/photo-deadbeefdeadbeef-100x50.png');
// strips a prior hashed suffix (re-derive doesn't stack)
eq('strips prior -<hash>-WxH', libraryOutputPath({ sourcePath: 'media/photo-aaaa1111bbbb2222-2048x1229.webp', fingerprint: 'cccc3333dddd4444', width: 800, height: 480, mime: 'image/webp' }), 'media/photo-cccc3333dddd4444-800x480.webp');

console.log('=== libraryFingerprint determinism / collision-resistance ===');
const fpA  = await libraryFingerprint({ file: fakeFile([1,2,3,4]),   sourceRect: RECT, options: OPTS });
const fpA2 = await libraryFingerprint({ file: fakeFile([1,2,3,4]),   sourceRect: RECT, options: OPTS });
ok('same source+crop+opts -> SAME fingerprint', fpA === fpA2);
const fpDiffSrc = await libraryFingerprint({ file: fakeFile([9,9,9,9]), sourceRect: RECT, options: OPTS });
ok('different source bytes -> DIFFERENT fingerprint', fpA !== fpDiffSrc);
const fpDiffCrop = await libraryFingerprint({ file: fakeFile([1,2,3,4]), sourceRect: { x: 100, y: 0, width: 4800, height: 3000 }, options: OPTS });
ok('different crop rect -> DIFFERENT fingerprint', fpA !== fpDiffCrop);
// quality "0.82" vs 0.82 must NOT diverge (normalised to Number)
const fpQStr = await libraryFingerprint({ file: fakeFile([1,2,3,4]), sourceRect: RECT, options: { ...OPTS, quality: '0.82' } });
ok('quality "0.82" == 0.82 (normalised)', fpA === fpQStr);
ok('fingerprint is long hex', /^[0-9a-f]{64}$/.test(fpA), fpA);

// end-to-end: two DIFFERENT sources with the SAME stem+dims produce DIFFERENT persisted paths
const pathA = libraryOutputPath({ sourcePath: 'media/logo.jpg', fingerprint: fpA,       width: 2048, height: 1229, mime: 'image/webp' });
const pathB = libraryOutputPath({ sourcePath: 'media/logo.jpg', fingerprint: fpDiffSrc, width: 2048, height: 1229, mime: 'image/webp' });
ok('two different sources, same stem+dims -> DIFFERENT persisted paths', pathA !== pathB);

console.log(`\n${pass} passed, ${fail} failed`);
if (typeof process !== 'undefined') process.exit(fail ? 1 : 0);
else if (typeof Deno !== 'undefined') Deno.exit(fail ? 1 : 0);
