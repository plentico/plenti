// Dependency-free regression tests for defaults/core/cms/crop-engine.js.
//
//   Run:  node scripts/test-crop-engine.mjs
//
// No test framework, no npm install. The engine is an ESM file that lives under
// defaults/core/ (embedded into the binary), and this repo has no root
// package.json marking .js as ESM, so we load the engine *source* through a
// data: URL import — which always parses as a module regardless of config.
//
// Covers the ADR test checklist (crop x scale matrix, format fallback, the
// null-selection regression, etc.). Canvas/FileReader are mocked just enough to
// drive the transformation paths.

import { readFile } from 'node:fs/promises';

let drawCall = null, fillRectCalled = false, forcePng = false;
const resetDraw = () => { drawCall = null; };
const resetFill = () => { fillRectCalled = false; };
const setForcePng = v => { forcePng = v; };

globalThis.document = {
    createElement() {
        const c = { width: 0, height: 0 };
        c.getContext = () => ({
            imageSmoothingEnabled: false, imageSmoothingQuality: '',
            set fillStyle(_) {}, get fillStyle() { return ''; },
            fillRect() { fillRectCalled = true; },
            drawImage(img, sx, sy, sw, sh, dx, dy, dw, dh) { drawCall = { sx, sy, sw, sh, dw, dh }; },
        });
        c.toBlob = (cb, format) => {
            let type = format;
            if (forcePng && (format === 'image/webp' || format === 'image/avif')) type = 'image/png';
            cb({ type, size: (c.width * c.height) || 1 });
        };
        c.toDataURL = format => `data:${format};base64,AAAA`;
        return c;
    },
};

const src = await readFile(new URL('../defaults/core/cms/crop-engine.js', import.meta.url), 'utf8');
const { sourceExtension, extToMime, parseImageOptions, renderImage, outputFilename, transformImage,
        conformsToImageOptions, LIBRARY_OPTIMISE_DEFAULTS } =
    await import('data:text/javascript,' + encodeURIComponent(src));

let pass = 0, fail = 0;
const eq = (actual, expected, name) => {
    const ok = JSON.stringify(actual) === JSON.stringify(expected);
    console.log(`${ok ? '✓' : '✗'} ${name}${ok ? '' : `\n    got      ${JSON.stringify(actual)}\n    expected ${JSON.stringify(expected)}`}`);
    ok ? pass++ : fail++;
};
const throwsAsync = async (fn, name) => {
    try { await fn(); console.log(`✗ ${name} (expected a throw)`); fail++; }
    catch { console.log(`✓ ${name} (threw as expected)`); pass++; }
};

const img = { naturalWidth: 800, naturalHeight: 600 };

console.log('--- pure logic ---');
eq(sourceExtension('media/perry.JPG?v=2'), 'jpg', 'sourceExtension');
eq([extToMime('webp'), extToMime('zzz')], ['image/webp', 'image/png'], 'extToMime');
eq(parseImageOptions({ hero: { type: 'media', options: [{ width: 500, height: 300, scale: true, crop: true, convert: 'webp' }] } }, 'hero'),
   { enabled: true, width: 500, height: 300, scale: true, crop: true, convert: 'webp', quality: 0.82, background: null, aspectRatio: '500:300', lockAspectRatio: false, minWidth: null, minHeight: null, showGrid: true, showCoordinates: true },
   'parseImageOptions maintainer options[]');
eq(parseImageOptions({ t: { type: 'media', crop: { enabled: true, aspectRatio: '1:1', minWidth: 200, minHeight: 200 } } }, 't'),
   { enabled: true, width: null, height: null, scale: true, crop: true, convert: null, quality: 0.82, background: null, aspectRatio: '1:1', lockAspectRatio: false, minWidth: 200, minHeight: 200, showGrid: true, showCoordinates: true },
   'parseImageOptions prototype crop{}');
eq(parseImageOptions({ x: { type: 'media' } }, 'x'), null, 'parseImageOptions no config -> null');
eq(parseImageOptions({ x: { type: 'media', options: [{ convert: 'gif' }] } }, 'x').convert, null, 'normalizeFormat rejects gif output');
eq(parseImageOptions({ x: { type: 'media', options: [{ convert: 'IMAGE/WEBP' }] } }, 'x').convert, 'webp', 'normalizeFormat anchored (no fooimage bug)');
eq(outputFilename('media/perry-cropped-1x1.webp', { width: 500, height: 300, ext: 'png' }), 'media/perry-500x300.png', 'outputFilename strips prior suffix');

console.log('\n--- renderImage / transformImage (ADR test checklist 1-10) ---');

// 1. null & undefined selection -> whole image
resetDraw(); let r = await renderImage(img, null, { format: 'image/png' });
eq([drawCall.sw, drawCall.sh, r.width, r.height], [800, 600, 800, 600], 'C1a null selection -> whole image');
resetDraw(); await renderImage(img, undefined, { format: 'image/png' });
eq([drawCall.sw, drawCall.sh], [800, 600], 'C1b undefined selection -> whole image');

// 2. selection at/beyond edge -> never zero dims
resetDraw(); await renderImage(img, { x: 800, y: 600, width: 100, height: 100 }, { format: 'image/png' });
eq([drawCall.sw >= 1, drawCall.sh >= 1], [true, true], 'C2 edge selection -> non-zero source rect');

// 3. scale:false preserves selected source-pixel dims
let res = await transformImage(img, { x: 10, y: 10, width: 400, height: 300 }, { crop: true, scale: false }, 'media/p.png');
eq([res.width, res.height], [400, 300], 'C3 scale:false preserves selection dims');

// 4. crop:false ignores selection, whole-image contain (800x600 within 500x500 = 500x375)
resetDraw(); res = await transformImage(img, { x: 10, y: 10, width: 400, height: 300 }, { crop: false, scale: true, width: 500, height: 500 }, 'media/p.png');
eq([drawCall.sw, drawCall.sh, res.width, res.height], [800, 600, 500, 375], 'C4 crop:false contain, ignores selection');

// 5. crop:true with no selection -> error
await throwsAsync(() => transformImage(img, null, { crop: true, scale: true, width: 400, height: 300 }, 'media/p.png'), 'C5 crop:true no selection throws');

// 6. fractional dims -> integers
res = await transformImage(img, { x: 0, y: 0, width: 400, height: 400 }, { crop: true, scale: true, width: 250.7, height: 250.7 }, 'media/p.png');
eq([Number.isInteger(res.width), res.width], [true, 251], 'C6 fractional dims -> integer');

// 7. aspect mismatch (both dims) -> rejected
await throwsAsync(() => transformImage(img, { x: 0, y: 0, width: 400, height: 200 }, { crop: true, scale: true, width: 300, height: 300 }, 'media/p.png'), 'C7 aspect mismatch rejected');

// 8. unsupported webp encode -> named PNG + formatFallback
setForcePng(true);
res = await transformImage(img, { x: 0, y: 0, width: 400, height: 300 }, { crop: true, scale: true, width: 400, height: 300, convert: 'webp' }, 'media/p.jpg');
eq([res.filePath, res.formatFallback], ['media/p-400x300.png', true], 'C8 webp fallback -> named PNG + flagged');
setForcePng(false);

// 9. .jpeg -> image/jpeg not marked converted
res = await transformImage(img, { x: 0, y: 0, width: 400, height: 300 }, { crop: true, scale: true, width: 400, height: 300, convert: 'jpg' }, 'media/photo.jpeg');
eq(res.converted, false, 'C9 .jpeg -> jpg not converted');

// 10. transparent -> JPEG gets background fill (webp does not)
resetFill(); await transformImage(img, { x: 0, y: 0, width: 400, height: 300 }, { crop: true, scale: true, width: 400, height: 300, convert: 'jpg' }, 'media/p.png');
eq(fillRectCalled, true, 'C10a JPEG output fills background');
resetFill(); await transformImage(img, { x: 0, y: 0, width: 400, height: 300 }, { crop: true, scale: true, width: 400, height: 300, convert: 'webp' }, 'media/p.png');
eq(fillRectCalled, false, 'C10b non-JPEG output does not fill');

// locked return contract (sourceRect added for the Media Library gateway fingerprint)
res = await transformImage(img, { x: 0, y: 0, width: 400, height: 300 }, { crop: true, scale: true, width: 400, height: 300 }, 'media/p.png');
eq(Object.keys(res), ['blob', 'filePath', 'width', 'height', 'sourceRect', 'requestedMime', 'actualMime', 'formatFallback', 'converted', 'bytes'], 'transformImage return shape is the frozen contract');
eq(res.bytes > 0, true, 'transformImage returns bytes (size-win proof)');
eq(res.sourceRect, { x: 0, y: 0, width: 400, height: 300 }, 'C-SR sourceRect is the normalised source rect');

// --- Media Library gateway: maxWidth/maxHeight (contain within a MAX edge, never a forced exact) ---
// C-MAX1: crop:false + maxW/H on an oversized whole image -> contain to the longest edge, keep aspect.
res = await transformImage(img, null, { crop: false, scale: true, maxWidth: 400, maxHeight: 400, convert: 'webp' }, 'media/p.png');
eq([res.width, res.height], [400, 300], 'C-MAX1 800x600 contained into max 400 -> 400x300 (not 400x400)');
// C-MAX2: crop:true + maxW/H (no exact dims) -> crop to selection, THEN contain the cropped rect.
res = await transformImage(img, { x: 0, y: 0, width: 800, height: 400 }, { crop: true, scale: true, maxWidth: 400, maxHeight: 400, convert: 'webp' }, 'media/p.png');
eq([res.width, res.height], [400, 200], 'C-MAX2 crop 800x400 then contain max 400 -> 400x200');
// C-MAX3: never upscale — a source already within the max stays its own size.
res = await transformImage(img, { x: 0, y: 0, width: 300, height: 200 }, { crop: true, scale: true, maxWidth: 2048, maxHeight: 2048, convert: 'webp' }, 'media/p.png');
eq([res.width, res.height], [300, 200], 'C-MAX3 within-max source is not upscaled');
// C-MAX4: mixing exact dims with max dims is rejected.
await throwsAsync(() => transformImage(img, { x: 0, y: 0, width: 400, height: 300 }, { crop: true, scale: true, width: 400, height: 300, maxWidth: 2048 }, 'media/p.png'), 'C-MAX4 exact + max dims rejected');

// --- conformance short-circuit (#364, maintainer-confirmed: reference conforming assets, no copy) ---
const conf = (w, h, path, opts) => conformsToImageOptions({ width: w, height: h, path }, opts);
// C-CONF1: crop:true + both dims -> exact match conforms; off-by-one does not.
eq(conf(500, 300, 'media/a.webp', { crop: true, scale: true, width: 500, height: 300, convert: 'webp' }), true, 'C-CONF1 exact dims + format conform');
eq(conf(500, 301, 'media/a.webp', { crop: true, scale: true, width: 500, height: 300, convert: 'webp' }), false, 'C-CONF1b off-by-one height does not conform');
// C-CONF2: crop:true without BOTH dims is underdetermined -> never conforms.
eq(conf(500, 300, 'media/a.webp', { crop: true, scale: true, width: 500, convert: 'webp' }), false, 'C-CONF2 one-dim crop never conforms');
eq(conf(500, 300, 'media/a.webp', { crop: true, scale: false }), false, 'C-CONF2b free crop never conforms');
// C-CONF3: crop:false + scale -> within bounds conforms (contain never upscales); over bounds does not.
eq(conf(400, 300, 'media/a.webp', { crop: false, scale: true, width: 500, height: 500, convert: 'webp' }), true, 'C-CONF3 within contain bounds conforms');
eq(conf(800, 300, 'media/a.webp', { crop: false, scale: true, width: 500, height: 500, convert: 'webp' }), false, 'C-CONF3b over a bound does not conform');
eq(conf(2048, 1229, 'media/a.webp', { crop: false, scale: true, maxWidth: 2048, maxHeight: 2048, convert: 'webp' }), true, 'C-CONF3c maxW/H bounds honoured (library defaults shape)');
// C-CONF4: format gate — convert set requires matching extension; jpg/jpeg are one format.
eq(conf(400, 300, 'media/a.png', { crop: false, scale: true, width: 500, height: 500, convert: 'webp' }), false, 'C-CONF4 wrong format does not conform');
eq(conf(400, 300, 'media/a.jpeg', { crop: false, scale: false, convert: 'jpg' }), true, 'C-CONF4b jpeg conforms to convert:jpg');
// C-ING: ingestion conformance against the LIBRARY defaults (the add-as-is gate,
// owner-confirmed): already-webp within the max edge is added byte-identical.
eq(conf(500, 300, 'media/500x300_Rubiaceae.webp', LIBRARY_OPTIMISE_DEFAULTS), true, 'C-ING1 conforming webp within bounds');
eq(conf(500, 300, 'media/photo.png', LIBRARY_OPTIMISE_DEFAULTS), false, 'C-ING2 wrong format never as-is');
eq(conf(2500, 1000, 'media/wide.webp', LIBRARY_OPTIMISE_DEFAULTS), false, 'C-ING3 over the max edge never as-is');
eq(conf(2048, 2048, 'media/edge.webp', LIBRARY_OPTIMISE_DEFAULTS), true, 'C-ING4 exactly at the max edge conforms');
eq(conf(400, 300, 'media/a.png', { crop: false, scale: false }), true, 'C-CONF4c no convert -> format always conforms');
// C-CONF5: scale:false -> format-only spec; missing dims -> never conforms.
eq(conf(9999, 9999, 'media/a.webp', { crop: false, scale: false, convert: 'webp' }), true, 'C-CONF5 scale:false ignores dimensions');
eq(conf(0, 300, 'media/a.webp', { crop: false, scale: false }), false, 'C-CONF5b unknown dims never conform');

console.log(`\n${pass} passed, ${fail} failed`);
process.exit(fail ? 1 : 0);
