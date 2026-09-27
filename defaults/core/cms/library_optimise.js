// library_optimise.js — Media Library upload GATEWAY helpers.
//
// Collision-resistant persisted identity for an optimised library derivative.
// The engine's outputFilename() names by source stem + output dims, which
// collides across different sources or crops that yield the same name
// (photo.jpg + photo.png -> both photo-2048x1365.webp). With upsert that would
// silently overwrite a live asset. So the gateway derives the persisted path
// from a HASH of the exact source bytes + normalised transform, keeping the
// engine's field-derivative naming untouched.
//
// Representation contract (see ADR 0001 D12): the returned path is the PERSISTED
// IDENTITY — the string stored in the Media Library grid and the field. It is
// never a data URL (that is transport) and never the engine's non-hashed
// result.filePath.

const MIME_TO_EXT = {
    'image/jpeg': 'jpg', 'image/png': 'png', 'image/webp': 'webp', 'image/avif': 'avif',
};

/** Extension for a mime (the gateway path carries the ACTUAL encoded format). */
export function extForMime(mime) {
    return MIME_TO_EXT[String(mime || '').toLowerCase()] || 'png';
}

/** Directory + stem of a source path, with any prior "-WxH"/"-cropped-WxH"/"-<hash>-WxH" stripped. */
function dirAndStem(srcPath) {
    const src = String(srcPath || '');
    const slash = src.lastIndexOf('/');
    const dir = src.slice(0, slash + 1);
    const base = src.slice(slash + 1);
    const dot = base.lastIndexOf('.');
    let stem = dot > -1 ? base.slice(0, dot) : base;
    // strip a trailing "-WxH" or "-cropped-WxH" or "-<hex>-WxH"
    stem = stem.replace(/-(?:cropped-)?(?:[0-9a-f]{6,}-)?\d+x\d+$/i, '');
    return { dir, stem };
}

/** Hex SHA-256 of an ArrayBuffer/Uint8Array via Web Crypto. */
async function sha256Hex(bytes) {
    if (!globalThis.crypto?.subtle) {
        throw new Error('Secure hashing is unavailable in this browser context (requires HTTPS or localhost).');
    }
    const buf = bytes instanceof ArrayBuffer ? bytes : bytes.buffer ?? bytes;
    const digest = await crypto.subtle.digest('SHA-256', buf);
    return [...new Uint8Array(digest)].map(b => b.toString(16).padStart(2, '0')).join('');
}

/** Stable JSON: sort object keys so equivalent payloads hash identically. */
function canonicalJSON(value) {
    if (value === null || typeof value !== 'object') return JSON.stringify(value);
    if (Array.isArray(value)) return '[' + value.map(canonicalJSON).join(',') + ']';
    const keys = Object.keys(value).sort();
    return '{' + keys.map(k => JSON.stringify(k) + ':' + canonicalJSON(value[k])).join(',') + '}';
}

/**
 * Deterministic fingerprint of "this exact source, cropped this exact way, with
 * these output settings". Same inputs -> same fingerprint (safe upsert update);
 * different source bytes OR different crop -> different fingerprint (no overwrite).
 * Two-stage: hash the source bytes, then hash a canonical payload that includes
 * that digest + the engine's OWN normalised sourceRect + the output options.
 * @param {{file: File|Blob, sourceRect: {x,y,width,height}, options: object}} args
 * @returns {Promise<string>} lowercase hex (full digest; the caller truncates)
 */
export async function libraryFingerprint({ file, sourceRect, options }) {
    const sourceBytes = await file.arrayBuffer();
    const sourceDigest = await sha256Hex(sourceBytes);
    const o = options || {};
    const rect = sourceRect || {};
    const payload = {
        version: 1,
        sourceDigest,
        sourceRect: {
            x: rect.x | 0, y: rect.y | 0,
            width: rect.width | 0, height: rect.height | 0,
        },
        maxWidth: o.maxWidth ?? null,
        maxHeight: o.maxHeight ?? null,
        format: o.convert ?? null,
        // normalise quality to a stable number so 0.82 and "0.82" never diverge
        quality: o.quality == null ? null : Number(o.quality),
    };
    return sha256Hex(new TextEncoder().encode(canonicalJSON(payload)));
}

/**
 * Build the collision-resistant persisted path for a library derivative:
 *   media/<stem>-<hash12>-<W>x<H>.<ext>
 * ext is derived from the ACTUAL encoded mime (so a webp->png fallback is named .png).
 * @param {{sourcePath: string, fingerprint: string, width: number, height: number, mime: string, hashLength?: number}} args
 */
export function libraryOutputPath({ sourcePath, fingerprint, width, height, mime, hashLength = 16 }) {
    const { dir, stem } = dirAndStem(sourcePath);
    const hash = String(fingerprint || '').slice(0, hashLength);
    const ext = extForMime(mime);
    return `${dir}${stem}-${hash}-${width}x${height}.${ext}`;
}
