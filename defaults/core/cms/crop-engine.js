// crop-engine.js — client-side image crop / scale / convert for the Plenti CMS.
//
// A dependency-free, framework-agnostic *transformation* module. It is NOT
// "pure": it depends on the browser Canvas/DOM API and attaches one guarded
// window global. It transforms only (crop / scale / convert + filename) and
// does NOT persist — the host shell submits the result through Plenti's
// existing provider pipeline (Button -> postLocal | commitGitlab | commitGitea).
// That keeps GitLab/Gitea/local saving, auth, and save-status UI in one place,
// and survives the Svelte -> Pico migration (cropping is client-side; that
// migration is server-side rendering — orthogonal).
//
// Delivery:
//   - today (Svelte CMS):  import { transformImage } from './crop-engine.js'
//   - later (Pico CMS):    <script type="module" src="/crop-engine.js"> + the
//                          window.PlentiImage global, for inline Pattr handlers.
//
// See docs/adr/0001-clientside-image-crop-scale-convert.md for the behaviour
// contract (the crop x scale matrix) and decisions.

const MIME_BY_EXT = {
    jpg: 'image/jpeg', jpeg: 'image/jpeg', png: 'image/png',
    webp: 'image/webp', avif: 'image/avif',
};
const EXT_BY_MIME = {
    'image/jpeg': 'jpg', 'image/png': 'png', 'image/webp': 'webp', 'image/avif': 'avif',
};
const OUTPUT_FORMATS = ['jpg', 'png', 'webp', 'avif']; // canvas encoders; gif is NOT encodable
const MAX_OUTPUT_PIXELS = 40000000; // 40 MP guard against runaway schema dimensions
const ASPECT_TOLERANCE = 0.01;      // 1% — tolerate integer-rounding, reject gross distortion

// Defaults for the Media Library upload GATEWAY (schema-free general optimisation): convert to a
// web-ready format and contain within a MAXIMUM edge (downscale oversized images, never upscale).
// crop:false = whole image by default; the standalone modal may toggle crop on. maxWidth/maxHeight
// are a MAX bound (contain), NOT a forced exact 2048x2048 — see transformImage's matrix.
export const LIBRARY_OPTIMISE_DEFAULTS = Object.freeze({
    crop: false,
    scale: true,
    convert: 'webp',
    quality: 0.82,
    maxWidth: 2048,
    maxHeight: 2048,
});

/** Lowercase file extension from a path/URL, without the dot. */
export function sourceExtension(path) {
    const m = String(path || '').toLowerCase().match(/\.([a-z0-9]+)(?:[?#].*)?$/);
    return m ? m[1] : '';
}

/** Map an extension to a canvas-encode mime type (defaults to image/png). */
export function extToMime(ext) {
    return MIME_BY_EXT[String(ext || '').toLowerCase()] || 'image/png';
}

/**
 * Read crop/scale/convert options off a content-type schema for one field.
 * Accepts BOTH shapes (issue #364 schema is still the maintainer's open call):
 *   - maintainer sketch: { type:"media", options:[{ width, height, scale, crop, convert }] }
 *   - prototype:         { type:"media", crop:{ enabled, aspectRatio, minWidth, ... } }
 * @returns normalized options, or null when the field has no crop config.
 */
export function parseImageOptions(schema, fieldKey) {
    if (!schema || !fieldKey) return null;
    const field = schema[fieldKey];
    if (!field || field.type !== 'media') return null;

    let raw = null;
    if (Array.isArray(field.options) && field.options.length) {
        raw = field.options[0]; // multi-option (responsive / art-direction) is future work
    } else if (field.crop && typeof field.crop === 'object') {
        raw = field.crop;
    } else {
        return null;
    }
    if (raw.enabled === false) return null;

    const width = positiveIntOrNull(raw.width != null ? raw.width : raw.outputWidth);
    const height = positiveIntOrNull(raw.height != null ? raw.height : raw.outputHeight);
    let aspectRatio = raw.aspectRatio || null;
    if (!aspectRatio && width && height) aspectRatio = `${width}:${height}`;

    return {
        enabled: true,
        width,
        height,
        scale: raw.scale !== false,                 // default: scale to dims when given
        crop: raw.crop !== false,                   // default: allow interactive crop
        convert: normalizeFormat(raw.convert),      // normalized output ext, or null = preserve source
        quality: clamp01(raw.quality != null ? normalizeQuality(raw.quality) : 0.82),
        background: typeof raw.background === 'string' ? raw.background : null,
        aspectRatio,
        lockAspectRatio: !!raw.lockAspectRatio,
        minWidth: positiveIntOrNull(raw.minWidth),
        minHeight: positiveIntOrNull(raw.minHeight),
        showGrid: raw.showGrid !== false,
        showCoordinates: raw.showCoordinates !== false,
    };
}

/**
 * Does an EXISTING asset already satisfy a field's image options — such that
 * selecting it should reference the original with NO derivative copy? (#364,
 * maintainer-confirmed: "if the image already meets the specifications …
 * just reference the existing image and not make a copy".)
 *
 * Deliberately conservative — true only when the spec fully determines the
 * outcome without user input:
 *  - format: when `convert` is set the source extension must already match
 *    (jpg/jpeg treated as one); no `convert` → format always conforms.
 *  - crop:true with BOTH width & height → exact dimension match required.
 *  - crop:true WITHOUT both dims (free/one-dim crop) → never conforms — the
 *    output depends on a user selection we cannot infer.
 *  - crop:false, scale:true → conforms when already within the configured
 *    bounds (contain semantics never upscale, so "within" means "unchanged").
 *  - crop:false, scale:false → dimensions always conform (format-only check).
 *
 * @param {{width:number, height:number, path:string}} asset natural dims + path
 * @param {?object} opts parseImageOptions() output (or a compatible shape)
 */
export function conformsToImageOptions(asset, opts) {
    if (!opts || !asset) return false;
    const w = positiveIntOrNull(asset.width);
    const h = positiveIntOrNull(asset.height);
    if (!w || !h) return false;

    if (opts.convert) {
        const dejpeg = e => (e === 'jpeg' ? 'jpg' : e);
        const want = dejpeg(normalizeFormat(opts.convert) || '');
        const have = dejpeg(sourceExtension(asset.path));
        if (!want || !have || want !== have) return false;
    }

    const cfgW = positiveIntOrNull(opts.width);
    const cfgH = positiveIntOrNull(opts.height);
    const maxW = positiveIntOrNull(opts.maxWidth);
    const maxH = positiveIntOrNull(opts.maxHeight);

    if (opts.crop !== false) {
        if (!(cfgW && cfgH)) return false;          // underdetermined crop — never short-circuit
        return w === cfgW && h === cfgH;
    }
    if (opts.scale === false) return true;          // format-only spec
    if (maxW || maxH) return (!maxW || w <= maxW) && (!maxH || h <= maxH);
    if (cfgW || cfgH) return (!cfgW || w <= cfgW) && (!cfgH || h <= cfgH);
    return true;                                    // scale:true but no bounds configured
}

/**
 * Low-level: crop + scale + convert in a single drawImage pass.
 * @param {HTMLImageElement} img loaded source (naturalWidth/Height).
 * @param {?{x,y,width,height}} selection source-pixel rect; null = whole image.
 * @param {{width?,height?,format?,quality?,background?}} output
 * @returns {Promise<{blob:Blob,mime:string,width:number,height:number}>}
 *          mime is what the browser ACTUALLY produced (may differ from requested,
 *          e.g. webp -> png on old Safari).
 */
export async function renderImage(img, selection, output = {}) {
    validateImage(img);
    const { x: sx, y: sy, width: sw, height: sh } = resolveSourceRect(img, selection);

    let dw = positiveIntOrNull(output.width);
    let dh = positiveIntOrNull(output.height);
    if (dw && !dh) dh = Math.max(1, Math.round(dw * sh / sw));
    else if (dh && !dw) dw = Math.max(1, Math.round(dh * sw / sh));
    else if (!dw && !dh) { dw = sw; dh = sh; }

    const canvas = document.createElement('canvas');
    canvas.width = dw;
    canvas.height = dh;
    const ctx = canvas.getContext('2d');
    if (!ctx) throw new Error('A 2D canvas context is unavailable.');
    ctx.imageSmoothingEnabled = true;
    ctx.imageSmoothingQuality = 'high';

    const format = output.format || 'image/png';
    if (format === 'image/jpeg') {
        // JPEG has no alpha — fill a background so transparency isn't rendered black.
        ctx.fillStyle = output.background || '#ffffff';
        ctx.fillRect(0, 0, dw, dh);
    }
    ctx.drawImage(img, sx, sy, sw, sh, 0, 0, dw, dh);

    const blob = await canvasToBlob(canvas, format, output.quality);
    return { blob, mime: blob.type || format, width: dw, height: dh };
}

/**
 * Build the output path: dir + stem + "-{W}x{H}." + ext, stripping any prior
 * "-WxH" / "-cropped-WxH" suffix so re-crops don't stack.
 * e.g. ("media/perry.jpg", {width:500,height:300,ext:"webp"}) -> "media/perry-500x300.webp"
 */
export function outputFilename(srcPath, opts) {
    const { width, height, ext } = opts || {};
    const src = String(srcPath || '');
    const slash = src.lastIndexOf('/');
    const dir = src.slice(0, slash + 1);
    const base = src.slice(slash + 1);
    const dot = base.lastIndexOf('.');
    let stem = dot > -1 ? base.slice(0, dot) : base;
    stem = stem.replace(/-(?:cropped-)?\d+x\d+$/, '');
    const e = String(ext || 'png').replace(/^\./, '');
    return `${dir}${stem}-${width}x${height}.${e}`;
}

/**
 * High-level: apply the crop x scale matrix (see ADR), render, and name the
 * file from what was ACTUALLY produced. Does NOT save — returns the encoded
 * result for the host shell to commit through the active provider.
 * @returns {Promise<{blob,filePath,width,height,requestedMime,actualMime,formatFallback,bytes,converted}>}
 */
export async function transformImage(img, selection, options, srcPath, _extra = {}) {
    validateImage(img);
    const o = options || {};
    const doCrop = o.crop !== false;
    const doScale = o.scale !== false;

    // 1. effective source rect
    let sel = null;
    if (doCrop) {
        if (!isValidSelection(selection)) {
            throw new Error('A crop selection is required when crop is enabled.');
        }
        sel = selection;
    }
    const rect = resolveSourceRect(img, sel);
    const srcW = rect.width, srcH = rect.height;

    if ((o.minWidth && srcW < o.minWidth) || (o.minHeight && srcH < o.minHeight)) {
        throw new Error(`Crop is below the minimum size (${o.minWidth || 0}x${o.minHeight || 0}).`);
    }

    // 2. output dimensions per matrix
    const cfgW = positiveIntOrNull(o.width);
    const cfgH = positiveIntOrNull(o.height);
    // maxWidth/maxHeight = "contain within a MAXIMUM edge" (the Media Library gateway's general
    // downscale), distinct from exact width/height. They are mutually exclusive with exact dims.
    const maxW = positiveIntOrNull(o.maxWidth);
    const maxH = positiveIntOrNull(o.maxHeight);
    if ((cfgW || cfgH) && (maxW || maxH)) {
        throw new Error('Use exact dimensions or maximum dimensions, not both.');
    }
    let outW, outH;
    if (!doScale) {
        outW = srcW; outH = srcH;                       // scale:false -> source-pixel size
    } else if (maxW || maxH) {
        // contain-after-crop: fit the (possibly cropped) rect within maxW/maxH, never upscale.
        const scale = containScale(srcW, srcH, maxW, maxH);
        outW = Math.max(1, Math.round(srcW * scale));
        outH = Math.max(1, Math.round(srcH * scale));
    } else if (doCrop) {
        if (cfgW && cfgH) {
            const ratioErr = Math.abs((srcW / srcH) - (cfgW / cfgH)) / (cfgW / cfgH);
            if (ratioErr > ASPECT_TOLERANCE) {
                throw new Error('Crop selection aspect ratio does not match the configured output dimensions.');
            }
            outW = cfgW; outH = cfgH;
        } else if (cfgW) { outW = cfgW; outH = Math.max(1, Math.round(cfgW * srcH / srcW)); }
        else if (cfgH) { outH = cfgH; outW = Math.max(1, Math.round(cfgH * srcW / srcH)); }
        else { outW = srcW; outH = srcH; }
    } else {
        const scale = containScale(srcW, srcH, cfgW, cfgH); // crop:false -> contain, never upscale
        outW = Math.max(1, Math.round(srcW * scale));
        outH = Math.max(1, Math.round(srcH * scale));
    }

    if (outW * outH > MAX_OUTPUT_PIXELS) {
        throw new Error('Requested output dimensions are too large.');
    }

    // 3. format (output set only; invalid/unencodable convert falls back to source then png)
    const requestedExt = normalizeFormat(o.convert) || sourceExtension(srcPath) || 'png';
    const requestedMime = extToMime(requestedExt);

    // 4. render
    const { blob, mime: actualMime, width, height } = await renderImage(img, rect, {
        width: outW, height: outH, format: requestedMime, quality: o.quality, background: o.background,
    });

    // 5. name from the ACTUAL output format; report conversion/fallback by MIME
    const actualExt = mimeToExt(actualMime) || requestedExt;
    const sourceMime = extToMime(sourceExtension(srcPath));
    return {
        blob,
        filePath: outputFilename(srcPath, { width, height, ext: actualExt }),
        width,
        height,
        // The engine's own normalised source rectangle (post-clamp/round). The Media Library
        // gateway fingerprints THESE exact pixels for its collision-resistant name rather than
        // re-deriving the rounding — so "same source pixels + same settings -> same identity".
        sourceRect: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
        requestedMime,
        actualMime,
        formatFallback: requestedMime !== actualMime,
        converted: actualMime !== sourceMime,
        bytes: blob.size,
    };
}

/** Read a Blob as a data URL (for the shell to build a {file, contents} commit item). */
export function blobToDataURL(blob) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(String(reader.result));
        reader.onerror = () => reject(reader.error || new Error('Failed to read blob'));
        reader.readAsDataURL(blob);
    });
}

// --- internal helpers -------------------------------------------------------

function validateImage(img) {
    if (!img) throw new TypeError('A source image is required.');
    const w = img.naturalWidth, h = img.naturalHeight;
    if (!Number.isFinite(w) || !Number.isFinite(h) || w <= 0 || h <= 0) {
        throw new Error('The source image has not loaded or has invalid dimensions.');
    }
}

function isValidSelection(sel) {
    return !!sel && num(sel.width, 0) > 0 && num(sel.height, 0) > 0;
}

function resolveSourceRect(img, selection) {
    const natW = img.naturalWidth, natH = img.naturalHeight;
    if (!isValidSelection(selection)) {
        return { x: 0, y: 0, width: natW, height: natH };
    }
    const x = clamp(Math.round(num(selection.x, 0)), 0, natW - 1);
    const y = clamp(Math.round(num(selection.y, 0)), 0, natH - 1);
    const width = clamp(Math.round(num(selection.width, natW - x)), 1, natW - x);
    const height = clamp(Math.round(num(selection.height, natH - y)), 1, natH - y);
    return { x, y, width, height };
}

function containScale(srcW, srcH, maxW, maxH) {
    let scale = 1;
    if (maxW) scale = Math.min(scale, maxW / srcW);
    if (maxH) scale = Math.min(scale, maxH / srcH);
    return Math.min(1, scale); // never upscale
}

function mimeToExt(mime) {
    return EXT_BY_MIME[String(mime || '').toLowerCase()] || '';
}

function normalizeFormat(value) {
    if (!value) return null;
    let f = String(value).toLowerCase().replace(/^image\//, '');
    if (f === 'jpeg') f = 'jpg';
    return OUTPUT_FORMATS.includes(f) ? f : null;
}

function canvasToBlob(canvas, format, quality) {
    return new Promise((resolve, reject) => {
        if (typeof canvas.toBlob === 'function') {
            canvas.toBlob(blob => {
                if (blob) return resolve(blob);
                try { resolve(dataURLToBlob(canvas.toDataURL(format, quality))); }
                catch (err) { reject(err); }
            }, format, quality);
        } else {
            try { resolve(dataURLToBlob(canvas.toDataURL(format, quality))); }
            catch (err) { reject(err); }
        }
    });
}

function dataURLToBlob(dataUrl) {
    const parts = String(dataUrl).split(',');
    const mime = (parts[0].match(/data:([^;]+)/) || [null, 'application/octet-stream'])[1];
    const bin = atob(parts[1] || '');
    const arr = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) arr[i] = bin.charCodeAt(i);
    return new Blob([arr], { type: mime });
}

function positiveIntOrNull(v) {
    if (v == null || v === '') return null;
    const n = Number(v);
    return Number.isFinite(n) && n > 0 ? Math.max(1, Math.round(n)) : null;
}
function num(v, fallback) {
    if (v == null || v === '') return fallback;
    const n = Number(v);
    return Number.isFinite(n) ? n : fallback;
}
function clamp(v, lo, hi) { return Math.min(hi, Math.max(lo, v)); }
function clamp01(v) { return Math.min(1, Math.max(0, v)); }
function normalizeQuality(q) {
    const n = Number(q);
    if (!Number.isFinite(n)) return 0.82;
    return n > 1 ? n / 100 : n; // accept 0-1 or 0-100
}

// Non-clobbering global for the Pico/Pattr world (inline handlers can't reach
// ESM imports). SSR-safe (guarded); harmless under a Svelte ESM import today.
// Intentionally only the high-level public API — the helpers stay ESM-only.
if (typeof window !== 'undefined') {
    window.PlentiImage = Object.assign(window.PlentiImage || {}, {
        parseImageOptions, transformImage, blobToDataURL,
    });
}
