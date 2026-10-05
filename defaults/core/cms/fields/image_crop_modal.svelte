<script>
    import { createEventDispatcher } from 'svelte';
    import { sourceExtension } from '../crop-engine.js';

    // Presentation + interaction only. The parent owns transformation, the
    // pending queue, the field value, and the source-path distinction.
    //   imageUrl : the source image to crop (the ORIGINAL, resolved by the parent)
    //   options  : normalized parseImageOptions() output for this field
    //   error    : message shown when the parent's transform fails (modal stays open)
    //
    // Two modes:
    //  - field mode (libraryMode=false): the field's placement crop — pan/zoom
    //    editor, "Apply crop"/"Cancel". Dispatches: confirm { image, selection,
    //    overrides } ; cancel.
    //  - library mode (the "Format image" flow): crop is ALWAYS enabled via a
    //    resizable crop area with handles, plus an output-size slider (which may
    //    upscale) and a format picker. Dispatches:
    //      next     { image, selection, overrides, state } — overrides is null
    //               when the user touched nothing (parent keeps the ORIGINAL);
    //               `state` is handed back as `initialState` when the user
    //               navigates back to this image.
    //      previous — step back one image (parent restores `state`).
    //      close    — the × button or a backdrop click.
    export let imageUrl;
    export let options = {};
    export let error = '';
    export let processing = false;
    export let libraryMode = false;
    // { index, total } (1-based) for the multi-image "Format image" run, or null.
    export let queuePosition = null;
    // Original filename — drives the "Leave original (.ext)" format label.
    export let sourceName = '';
    // Prior edits for this image when navigating back: { rect, pct, format }.
    export let initialState = null;

    const dispatch = createEventDispatcher();
    const MAX = 360;      // field mode: max crop-box display edge (px)
    const STAGE = 460;    // library mode: max stage display edge (px)
    const MIN_CROP = 8;   // library mode: min crop rect edge (display px)
    const MAX_OUT = 4096; // library mode: max output edge (px)
    const HANDLES = ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w'];

    // ---------------- field mode (placement crop, pan/zoom) ----------------
    $: opts = options;
    $: doCrop = opts?.crop !== false;
    $: doScale = opts?.scale !== false;
    $: showGrid = opts?.showGrid !== false;
    $: showCoords = opts?.showCoordinates !== false;
    $: locked = !!(opts?.width && opts?.height) || (!!opts?.aspectRatio && !!opts?.lockAspectRatio);

    function aspectOf(o) {
        if (!o) return null;
        if (o.width && o.height) return o.width / o.height;
        if (o.aspectRatio) {
            const [w, h] = String(o.aspectRatio).split(':').map(Number);
            if (w > 0 && h > 0) return w / h;
        }
        return null;
    }
    $: aspect = aspectOf(opts);
    // Crop box: matches the locked output aspect; square for free crop; square
    // preview frame when not cropping (image is contained inside).
    $: cropSize = doCrop && aspect
        ? (aspect >= 1 ? { width: MAX, height: Math.round(MAX / aspect) }
                       : { width: Math.round(MAX * aspect), height: MAX })
        : { width: MAX, height: MAX };

    // Display state. `natural` is the image's pixel size; `fit` is the scale that
    // makes the image cover the crop box; `scale` (relative to fit) and `pos`
    // (top-left offset in box px) are the user's zoom/pan.
    let natural = null;
    let fitScale = 1;
    let scale = 1;
    let pos = { x: 0, y: 0 };
    let imageElement = null;

    function fit() {
        natural = imageElement ? { w: imageElement.naturalWidth, h: imageElement.naturalHeight } : null;
        if (!natural || !natural.w || !natural.h) return;
        fitScale = Math.max(cropSize.width / natural.w, cropSize.height / natural.h);
        scale = 1;
        center();
    }
    function center() {
        if (!natural) return;
        const w = natural.w * fitScale * scale, h = natural.h * fitScale * scale;
        pos = { x: (cropSize.width - w) / 2, y: (cropSize.height - h) / 2 };
    }
    function clampPos() {
        if (!natural) return;
        const w = natural.w * fitScale * scale, h = natural.h * fitScale * scale;
        pos = {
            x: Math.min(0, Math.max(cropSize.width - w, pos.x)),
            y: Math.min(0, Math.max(cropSize.height - h, pos.y)),
        };
    }

    // Crop selection in SOURCE pixels: the crop-box window mapped back through
    // the display transform. The crop box is the fixed window; pan/zoom moves
    // the image behind it.
    $: sel = natural
        ? {
            x: Math.min(natural.w - 1, Math.max(0, Math.round(-pos.x / (fitScale * scale)))),
            y: Math.min(natural.h - 1, Math.max(0, Math.round(-pos.y / (fitScale * scale)))),
            width: Math.max(1, Math.round(cropSize.width / (fitScale * scale))),
            height: Math.max(1, Math.round(cropSize.height / (fitScale * scale))),
        }
        : null;

    // Output dims when the caller scales the selection (see transformImage):
    // width&height = exact; one dim = proportional; maxWidth/maxHeight = contain
    // within the bound (never upscale); else the selection's own pixel size.
    $: outputDims = sel ? (() => {
        if (doScale) {
            if (opts?.width && opts?.height) return { width: opts.width, height: opts.height };
            if (opts?.width) return { width: opts.width, height: Math.round(opts.width * sel.height / sel.width) };
            if (opts?.height) return { width: Math.round(opts.height * sel.width / sel.height), height: opts.height };
            if (opts?.maxWidth || opts?.maxHeight) {
                const s = Math.min(1,
                    opts.maxWidth ? opts.maxWidth / sel.width : 1,
                    opts.maxHeight ? opts.maxHeight / sel.height : 1);
                return { width: Math.max(1, Math.round(sel.width * s)), height: Math.max(1, Math.round(sel.height * s)) };
            }
        }
        return { width: sel.width, height: sel.height };
    })() : null;

    let dragging = false, last = null;
    function down(e) {
        if (!doCrop) return;
        dragging = true; last = { x: e.clientX, y: e.clientY };
        window.addEventListener('mousemove', move);
        window.addEventListener('mouseup', up);
    }
    function move(e) {
        if (!dragging) return;
        pos = { x: pos.x + e.clientX - last.x, y: pos.y + e.clientY - last.y };
        last = { x: e.clientX, y: e.clientY };
        clampPos();
    }
    function up() {
        dragging = false;
        window.removeEventListener('mousemove', move);
        window.removeEventListener('mouseup', up);
    }

    function zoom(f) { scale = Math.min(8, Math.max(1, scale * f)); clampPos(); }

    function confirm() {
        if (!sel) return;
        // Schema-locked dims imply an exact-size derivative — without an explicit
        // schema convert the field opted into resizing, so re-encode to web-ready
        // jpg rather than emitting e.g. a 10MB PNG at exact dims.
        const overrides = {
            maxWidth: opts?.maxWidth ?? null,
            maxHeight: opts?.maxHeight ?? null,
            convert: (opts?.width && opts?.height) ? (opts.convert || 'jpg') : (opts?.convert ?? null),
            quality: opts?.quality,
        };
        if (doScale && opts?.width) overrides.width = opts.width;
        if (doScale && opts?.height) overrides.height = opts.height;
        dispatch('confirm', { image: imageElement, selection: doCrop ? sel : null, overrides });
    }
    function cancel() { dispatch('cancel'); }

    // ------------- library mode ("Format image": handles + output size) -------------
    let libNatural = null;  // { w, h } source px
    let libScale = 1;       // display px per source px (may upscale small images)
    let dispW = 0, dispH = 0;
    let rect = null;        // crop area in DISPLAY coords { x, y, w, h }
    let pct = 100;          // output size as % of the crop's native pixel size
    let format = '';        // '' = leave original
    let touched = false;    // any user edit; untouched keeps the ORIGINAL image
    let drag = null;        // { mode, startX, startY, rect0 } while dragging

    function fitLib() {
        if (!imageElement) return;
        const w = imageElement.naturalWidth, h = imageElement.naturalHeight;
        if (!w || !h) return;
        libNatural = { w, h };
        libScale = Math.min(STAGE / w, STAGE / h); // no cap: small images display larger
        dispW = Math.round(w * libScale);
        dispH = Math.round(h * libScale);
        if (initialState?.rect) {
            // Navigating back to a previously edited image — restore, and stay
            // "touched" so re-confirming regenerates the same derivative.
            rect = { ...initialState.rect };
            pct = initialState.pct ?? 100;
            format = initialState.format ?? '';
            touched = true;
        } else {
            rect = { x: 0, y: 0, w: dispW, h: dispH };
            pct = 100;
            format = '';
            touched = false;
        }
    }

    // Crop selection in SOURCE pixels (display rect mapped back by libScale).
    $: libSel = (() => {
        if (!rect || !libNatural) return null;
        const x = Math.min(libNatural.w - 1, Math.max(0, Math.round(rect.x / libScale)));
        const y = Math.min(libNatural.h - 1, Math.max(0, Math.round(rect.y / libScale)));
        return {
            x, y,
            width: Math.min(libNatural.w - x, Math.max(1, Math.round(rect.w / libScale))),
            height: Math.min(libNatural.h - y, Math.max(1, Math.round(rect.h / libScale))),
        };
    })();

    // Output size: pct scales the crop's native pixels — over 100% UPSCALES a
    // small crop to a larger output (exact width/height engine semantics).
    $: outW = libSel ? Math.min(MAX_OUT, Math.max(1, Math.round(libSel.width * pct / 100))) : 0;
    $: outH = libSel && outW ? Math.min(MAX_OUT, Math.max(1, Math.round(outW * libSel.height / libSel.width))) : 0;

    const dejpeg = e => (e === 'jpeg' ? 'jpg' : e);
    $: sourceExt = sourceExtension(sourceName);
    $: convertOptions = ['webp', 'png', 'jpg'].filter(f => f !== dejpeg(sourceExt));

    function startMove(e) {
        if (processing || !rect) return;
        drag = { mode: 'move', startX: e.clientX, startY: e.clientY, rect0: { ...rect } };
    }
    function startResize(e, handle) {
        if (processing || !rect) return;
        drag = { mode: handle, startX: e.clientX, startY: e.clientY, rect0: { ...rect } };
    }
    function onDragMove(e) {
        if (!drag || !rect) return;
        const dx = e.clientX - drag.startX, dy = e.clientY - drag.startY;
        const r = drag.rect0, m = drag.mode;
        let { x, y, w, h } = r;
        if (m === 'move') {
            x = Math.min(dispW - w, Math.max(0, r.x + dx));
            y = Math.min(dispH - h, Math.max(0, r.y + dy));
        } else {
            if (m.includes('e')) w = r.w + dx;
            if (m.includes('s')) h = r.h + dy;
            if (m.includes('w')) { x = r.x + dx; w = r.w - dx; }
            if (m.includes('n')) { y = r.y + dy; h = r.h - dy; }
            if (w < MIN_CROP) { if (m.includes('w')) x -= MIN_CROP - w; w = MIN_CROP; }
            if (h < MIN_CROP) { if (m.includes('n')) y -= MIN_CROP - h; h = MIN_CROP; }
            x = Math.max(0, Math.min(x, dispW - MIN_CROP));
            y = Math.max(0, Math.min(y, dispH - MIN_CROP));
            w = Math.min(w, dispW - x);
            h = Math.min(h, dispH - y);
        }
        touched = true;
        rect = { x, y, w, h };
    }
    function onDragEnd() { drag = null; }

    function next() {
        if (!libSel) return;
        dispatch('next', {
            image: imageElement,
            selection: touched ? libSel : null,
            overrides: touched ? {
                crop: true,
                width: outW,
                height: outH,
                convert: format || null,
                maxWidth: null,
                maxHeight: null,
            } : null,
            state: { rect: rect ? { ...rect } : null, pct, format },
        });
    }
    function previous() { dispatch('previous'); }
    function close() { if (!processing) dispatch('close'); }

    // Render at <body> level so the fixed overlay escapes the CMS edit-tray's
    // transform (which would otherwise become its containing block / clip it).
    function portal(node) {
        document.body.appendChild(node);
        return { destroy() { if (node.parentNode) node.parentNode.removeChild(node); } };
    }
</script>

<svelte:window on:mousemove={onDragMove} on:mouseup={onDragEnd} />

<div class="crop-modal" use:portal on:mousedown|self={libraryMode ? close : cancel}>
    <div class="panel">
        {#if libraryMode}
            <button type="button" class="close-x" on:click|preventDefault={close} disabled={processing} aria-label="Close">×</button>
            <h3>Format image</h3>
            {#if queuePosition}
                <p class="queue-pos">Image {queuePosition.index} of {queuePosition.total}</p>
            {/if}
            <p class="hint">Drag the crop area or its handles to set the crop.<br>(touching nothing keeps the original image)</p>
            <div class="lib-stage" style="width:{dispW}px; height:{dispH}px;">
                <img class="lib-src" bind:this={imageElement} src={imageUrl} alt="" draggable="false"
                     on:load={fitLib} style="width:{dispW}px; height:{dispH}px;" />
                {#if rect}
                    <div class="lib-crop" class:moving={drag && drag.mode === 'move'}
                         style="left:{rect.x}px; top:{rect.y}px; width:{rect.w}px; height:{rect.h}px;"
                         on:mousedown|preventDefault={startMove}>
                        <div class="grid"><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i></div>
                        {#each HANDLES as handle}
                            <span class="handle {handle}"
                                  on:mousedown|stopPropagation|preventDefault={e => startResize(e, handle)}></span>
                        {/each}
                    </div>
                {/if}
            </div>
            {#if libSel}
                <div class="readout">Output: {outW}×{outH}px</div>
            {/if}
            <div class="size-row">
                <span>Output size</span>
                <input type="range" min="10" max="400" step="5" bind:value={pct}
                       on:input={() => (touched = true)} disabled={processing} />
                <span class="pct">{pct}%</span>
                <button type="button" class="reset" on:click|preventDefault={fitLib} disabled={processing}>Reset</button>
            </div>
            <div class="optimize">
                <span class="opt-title">Optimize</span>
                <label>
                    <input type="radio" name="format" bind:group={format} value=""
                           on:change={() => (touched = true)} disabled={processing} />
                    Leave original{sourceExt ? ` (.${sourceExt})` : ''}
                </label>
                {#each convertOptions as f}
                    <label class:recommended={f === 'webp'}>
                        <input type="radio" name="format" bind:group={format} value={f}
                               on:change={() => (touched = true)} disabled={processing} />
                        <span class="opt-label">Convert to .{f}{#if f === 'webp'}<em class="badge">recommended</em>{/if}</span>
                    </label>
                {/each}
            </div>
            {#if error}
                <div class="err">{error}</div>
            {/if}
            <div class="actions">
                <button type="button" class="secondary" on:click|preventDefault={previous}
                        disabled={processing || !queuePosition || queuePosition.index <= 1}>Previous</button>
                <button type="button" class="primary" on:click|preventDefault={next} disabled={processing}>
                    {processing ? 'Processing…' : 'Next'}
                </button>
            </div>
        {:else}
            <h3>{doCrop ? 'Crop image' : 'Scale image'}</h3>
            <p class="hint">{doCrop ? 'Drag to position • use zoom to scale' : 'Preview of the scaled image'}</p>
            <div class="stage" class:dragging
                 style="width:{cropSize.width}px; height:{cropSize.height}px;"
                 on:mousedown|preventDefault={down}>
                <img class="src" bind:this={imageElement} src={imageUrl} alt="" draggable="false" on:load={fit}
                     style="width:{natural ? natural.w * fitScale * scale : 0}px; height:{natural ? natural.h * fitScale * scale : 0}px; transform: translate({pos.x}px, {pos.y}px);" />
                {#if doCrop && showGrid}
                    <div class="grid"><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i></div>
                {/if}
            </div>
            {#if showCoords && sel}
                <div class="readout">
                    {doCrop ? `Crop: ${sel.x}, ${sel.y} • ${sel.width}×${sel.height}px` : 'Whole image'}
                    {outputDims ? ` → Output: ${outputDims.width}×${outputDims.height}px` : ''}
                </div>
            {/if}
            {#if doCrop}
                <div class="zoom">
                    <button type="button" on:click|preventDefault={() => zoom(1 / 1.2)} disabled={processing}>−</button>
                    <span>{Math.round(scale * 100)}%</span>
                    <button type="button" on:click|preventDefault={() => zoom(1.2)} disabled={processing}>+</button>
                    <button type="button" class="reset" on:click|preventDefault={fit} disabled={processing}>Reset</button>
                </div>
            {/if}
            {#if error}
                <div class="err">{error}</div>
            {/if}
            <div class="actions">
                <button type="button" class="secondary" on:click|preventDefault={cancel} disabled={processing}>Cancel</button>
                <button type="button" class="primary" on:click|preventDefault={confirm} disabled={processing}>
                    {processing ? 'Processing…' : (doCrop ? 'Apply crop' : 'Apply')}
                </button>
            </div>
        {/if}
    </div>
</div>

<style>
    .crop-modal {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, .6);
        display: flex;
        align-items: center;
        justify-content: center;
        /* Above the Media modal wrapper (.plenti-modal-wrapper, z-index 99999): this
           modal is launched FROM the Media library, so it must stack over it. */
        z-index: 100000;
    }
    .panel {
        position: relative;
        background: #fff;
        border-radius: 8px;
        padding: 20px;
        max-width: 90vw;
        max-height: calc(100vh - 40px);
        overflow-y: auto;
        box-sizing: border-box;
        box-shadow: 0 10px 40px rgba(0, 0, 0, .3);
        text-align: center;
    }
    .close-x {
        position: absolute;
        top: 8px;
        right: 8px;
        width: 28px;
        height: 28px;
        border: none;
        border-radius: 50%;
        background: #e7e7e7;
        color: #333;
        font-size: 1.1rem;
        line-height: 1;
        cursor: pointer;
    }
    .close-x:hover { background: #d5d5d5; }
    h3 { margin: 0 0 4px; }
    .queue-pos { margin: 0 0 4px; color: #1c7fc7; font-size: .8rem; font-weight: bold; }
    .hint { margin: 0 0 14px; color: #666; font-size: .85rem; }
    .stage {
        position: relative;
        margin: 0 auto;
        overflow: hidden;
        border: 2px solid #1c7fc7;
        background: #f4f4f4 repeating-conic-gradient(#e9e9e9 0% 25%, #fff 0% 50%) 0 / 20px 20px;
        cursor: grab;
    }
    .stage.dragging { cursor: grabbing; }
    .src { position: absolute; top: 0; left: 0; user-select: none; pointer-events: none; max-width: none; }
    .lib-stage {
        position: relative;
        margin: 0 auto;
        overflow: hidden;
        border: 2px solid #1c7fc7;
        background: #f4f4f4 repeating-conic-gradient(#e9e9e9 0% 25%, #fff 0% 50%) 0 / 20px 20px;
        user-select: none;
    }
    .lib-src { position: absolute; top: 0; left: 0; pointer-events: none; }
    .lib-crop {
        position: absolute;
        box-sizing: border-box;
        border: 2px solid #1c7fc7;
        /* Dim everything OUTSIDE the crop area. */
        box-shadow: 0 0 0 9999px rgba(0, 0, 0, .45);
        cursor: move;
    }
    .lib-crop.moving { cursor: grabbing; }
    .handle {
        position: absolute;
        width: 12px;
        height: 12px;
        box-sizing: border-box;
        background: #fff;
        border: 2px solid #1c7fc7;
        border-radius: 2px;
    }
    .handle.nw { top: -7px; left: -7px; cursor: nwse-resize; }
    .handle.n  { top: -7px; left: calc(50% - 6px); cursor: ns-resize; }
    .handle.ne { top: -7px; right: -7px; cursor: nesw-resize; }
    .handle.e  { top: calc(50% - 6px); right: -7px; cursor: ew-resize; }
    .handle.se { bottom: -7px; right: -7px; cursor: nwse-resize; }
    .handle.s  { bottom: -7px; left: calc(50% - 6px); cursor: ns-resize; }
    .handle.sw { bottom: -7px; left: -7px; cursor: nesw-resize; }
    .handle.w  { top: calc(50% - 6px); left: -7px; cursor: ew-resize; }
    .grid {
        position: absolute;
        inset: 0;
        display: grid;
        grid-template: repeat(3, 1fr) / repeat(3, 1fr);
        pointer-events: none;
    }
    .grid i { border: 1px solid rgba(255, 255, 255, .4); }
    .readout { margin: 10px 0 0; font-size: .8rem; color: #333; }
    .zoom { display: flex; align-items: center; justify-content: center; gap: 8px; margin: 12px 0; }
    .zoom button { width: 32px; height: 32px; border: 1px solid #ccc; border-radius: 4px; background: #fff; cursor: pointer; }
    .zoom .reset { width: auto; padding: 0 12px; }
    .zoom span { min-width: 48px; }
    .size-row { display: flex; align-items: center; justify-content: center; gap: 8px; margin: 12px 0 0; font-size: .9rem; }
    .size-row input[type="range"] { flex: 0 1 220px; }
    .size-row .pct { min-width: 42px; }
    .size-row .reset { padding: 4px 12px; border: 1px solid #ccc; border-radius: 4px; background: #fff; cursor: pointer; }
    .optimize {
        margin: 25px auto 0;
        padding: 10px 12px;
        border: 1px solid #dfe1e8;
        border-radius: 8px;
        text-align: left;
        font-size: .9rem;
        background: #f8f9fb;
    }
    .optimize .opt-title { display: block; font-weight: bold; margin-bottom: 6px; }
    .optimize label {
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 4px 0;
        padding: 5px 7px;
        border: 1px solid transparent;
        border-radius: 6px;
        cursor: pointer;
    }
    .optimize label:hover { background: #eef1f6; }
    .optimize input { margin: 0; cursor: pointer; accent-color: #1c7fc7; }
    .optimize .opt-label { flex: 1; display: flex; align-items: center; }
    .optimize label.recommended {
        background: #eaf3fd;
        border-color: #bcd6f7;
    }
    .optimize label.recommended:hover { background: #e0eefb; }
    .optimize .badge {
        font-style: normal;
        font-size: .72rem;
        font-weight: bold;
        letter-spacing: .02em;
        color: #fff;
        background: #1c7fc7;
        border-radius: 10px;
        padding: 1px 8px;
        margin-left: 8px;
    }
    .err { color: darkred; margin: 8px 0; font-size: .85rem; }
    .actions { display: flex; gap: 10px; margin-top: 14px; }
    .actions button { flex: 1; padding: 10px; border: none; border-radius: 6px; font-weight: bold; cursor: pointer; }
    .actions .primary { background: #1c7fc7; color: #fff; }
    .actions .secondary { background: #e7e7e7; }
    .actions button:disabled { opacity: .5; cursor: default; }
</style>

