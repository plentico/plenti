<script>
    import { createEventDispatcher } from 'svelte';

    // Presentation + interaction only. The parent owns transformation, the
    // pending queue, the field value, and the source-path distinction.
    //   imageUrl : the source image to crop (the ORIGINAL, resolved by the parent)
    //   options  : normalized parseImageOptions() output for this field
    //   error    : message shown when the parent's transform fails (modal stays open)
    // Dispatches: confirm { image, selection } ; cancel
    export let imageUrl;
    export let options = {};
    export let error = '';
    // Display-only failures elsewhere in the batch; removal belongs to the queue.
    export let batchFailures = [];
    export let processing = false;
    // Library-gateway mode: shows the optimise preview by default with an OPTIONAL
    // "Crop image" toggle (standalone only — field-launched keeps allowCropToggle
    // false so the field's own placement crop is the single interactive step).
    export let libraryMode = false;
    export let allowCropToggle = false;
    // Ingestion conformance (#364): the source ALREADY meets the library
    // defaults (target format, within the max edge). Confirming without crop
    // adds the ORIGINAL BYTES — no decode/re-encode — so the copy reflects it.
    export let conforming = false;
    export let confirmLabel = '';
    // Standalone multi-file queue affordances (field call sites leave these off →
    // pixel-identical to before). `queuePosition` = { index, total } (1-based) or null
    // — DISPLAY data only; `cancelLabel` overrides the primary Cancel text; a
    // "Skip remaining" button shows only when `showCancelAll` is set (dispatches
    // 'cancelAll'). `queueMode` is the EXPLICIT behavioural flag (never inferred
    // from queuePosition): in queue mode the backdrop is inert — cancelling a
    // queued file is destructive and must be an explicit button press.
    export let queuePosition = null;
    export let cancelLabel = 'Cancel';
    export let showCancelAll = false;
    export let queueMode = false;

    const dispatch = createEventDispatcher();
    const MAX = 360; // max crop-box display edge (px)

    // In library mode the user may toggle cropping; wantCrop overrides options.crop.
    let wantCrop = false;
    // Effective options everything downstream reads (doCrop, cropSize, outputDims,
    // the confirm overrides). In library mode crop follows the toggle; elsewhere it
    // is exactly the incoming options.
    $: opts = libraryMode ? { ...options, crop: wantCrop } : options;

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

    let imageElement;
    let natural = { w: 0, h: 0 };
    let scale = 1;
    let pos = { x: 0, y: 0 };
    let isDragging = false, dragStart = { x: 0, y: 0 }, startPos = { x: 0, y: 0 };

    // Centre the source: cover the crop box (crop mode) or contain it (preview).
    function fit() {
        if (!imageElement) return;
        natural = { w: imageElement.naturalWidth, h: imageElement.naturalHeight };
        if (!natural.w || !natural.h) return;
        scale = doCrop
            ? Math.max(cropSize.width / natural.w, cropSize.height / natural.h)
            : Math.min(cropSize.width / natural.w, cropSize.height / natural.h);
        pos = { x: (cropSize.width - natural.w * scale) / 2, y: (cropSize.height - natural.h * scale) / 2 };
    }

    function startDrag(e) {
        if (!doCrop) return;
        e.preventDefault();
        isDragging = true;
        dragStart = { x: e.clientX, y: e.clientY };
        startPos = { ...pos };
    }
    function onMove(e) {
        if (!isDragging) return;
        pos = { x: startPos.x + (e.clientX - dragStart.x), y: startPos.y + (e.clientY - dragStart.y) };
    }
    function onUp() { isDragging = false; }
    function zoom(delta) {
        if (!doCrop) return;
        const next = Math.max(0.05, Math.min(scale + delta, 20));
        const k = next / scale;
        const cx = cropSize.width / 2, cy = cropSize.height / 2;
        pos = { x: cx - (cx - pos.x) * k, y: cy - (cy - pos.y) * k };
        scale = next;
    }

    // Source-pixel rect currently framed by the crop box (null = whole image).
    function selectionRect() {
        if (!doCrop || !natural.w) return null;
        const sx = Math.max(0, Math.round(-pos.x / scale));
        const sy = Math.max(0, Math.round(-pos.y / scale));
        const sw = Math.min(Math.round(cropSize.width / scale), natural.w - sx);
        const sh = Math.min(Math.round(cropSize.height / scale), natural.h - sy);
        return { x: sx, y: sy, width: Math.max(1, sw), height: Math.max(1, sh) };
    }
    // Reactive: recompute when pan/zoom/size change (deps referenced explicitly).
    $: sel = (pos, scale, natural, cropSize, imageElement ? selectionRect() : null);

    // Displayed output dimensions — MUST mirror transformImage's matrix (crop-engine.js)
    // so the readout matches what actually gets saved. Reads `opts` (the effective
    // options), and — the fix — honours maxWidth/maxHeight (contain within a MAX edge),
    // so an oversized source shows e.g. 2048×1229, not the source dims or a forced square.
    function outputDims(s) {
        const cfgW = opts?.width, cfgH = opts?.height;
        const maxW = opts?.maxWidth, maxH = opts?.maxHeight;
        // The source-pixel rect we're producing from: the crop selection, else whole image.
        const rect = (doCrop && s) ? s : { width: natural.w, height: natural.h };
        if (!doScale) return { w: rect.width, h: rect.height };
        if (maxW || maxH) {
            // contain the rect within the max edge(s), never upscale
            let k = 1;
            if (maxW) k = Math.min(k, maxW / rect.width);
            if (maxH) k = Math.min(k, maxH / rect.height);
            k = Math.min(1, k);
            return { w: Math.max(1, Math.round(rect.width * k)), h: Math.max(1, Math.round(rect.height * k)) };
        }
        if (doCrop) {
            if (cfgW && cfgH) return { w: cfgW, h: cfgH };
            if (cfgW) return { w: cfgW, h: Math.round(cfgW * rect.height / rect.width) };
            if (cfgH) return { h: cfgH, w: Math.round(cfgH * rect.width / rect.height) };
            return { w: rect.width, h: rect.height };
        }
        // crop:false, scale:true → contain whole image within cfgW/cfgH
        let k = 1;
        if (cfgW) k = Math.min(k, cfgW / natural.w);
        if (cfgH) k = Math.min(k, cfgH / natural.h);
        k = Math.min(1, k);
        return { w: Math.max(1, Math.round(natural.w * k)), h: Math.max(1, Math.round(natural.h * k)) };
    }
    $: dims = natural.w ? outputDims(sel) : null;

    // In library mode, fold the crop toggle into `overrides` the parent merges over
    // LIBRARY_OPTIMISE_DEFAULTS. crop:false selection is null (whole-image optimise).
    const confirm = () => dispatch('confirm', {
        image: imageElement,
        selection: doCrop ? selectionRect() : null,
        overrides: libraryMode ? { crop: wantCrop } : undefined,
    });
    const cancel = () => dispatch('cancel');
    const cancelAll = () => dispatch('cancelAll');

    // Render at <body> level so the fixed overlay escapes the CMS edit-tray's
    // transform (which would otherwise become its containing block / clip it).
    function portal(node) {
        document.body.appendChild(node);
        return { destroy() { if (node.parentNode) node.parentNode.removeChild(node); } };
    }
</script>

<svelte:window on:mousemove={onMove} on:mouseup={onUp} />

<!-- Backdrop dismiss: inert in queue mode (an accidental click must never skip a
     file) and inert while processing (a cancel must never race an in-flight
     commit/transform) — only field mode, at rest, treats it as a plain cancel. -->
<div class="crop-modal" use:portal on:mousedown|self={() => { if (!processing && !queueMode) cancel(); }}>
    <div class="panel">
        <h3>{doCrop ? 'Crop image' : 'Optimise image'}</h3>
        {#if queuePosition && queuePosition.total > 1}
            <p class="queue-pos">Image {queuePosition.index} of {queuePosition.total}</p>
        {/if}
        <p class="hint">{doCrop ? 'Drag to pan • scroll or buttons to zoom'
            : (conforming ? 'Already optimised — will be added unchanged' : 'Preview of the optimised output')}</p>

        {#if batchFailures.length}
            <div class="batch-warning" role="status">
                <strong>{batchFailures.length} file{batchFailures.length === 1 ? '' : 's'} could not be processed.</strong>
                <ul>{#each batchFailures as failure}
                    <li>{failure.name}: {failure.message}</li>
                {/each}</ul>
                <p>Finish or skip this image, then remove failed files before saving.</p>
            </div>
        {/if}

        {#if allowCropToggle}
            <label class="crop-toggle">
                <input type="checkbox" bind:checked={wantCrop} />
                Crop image
            </label>
        {/if}

        <div class="stage" class:dragging={isDragging} style="width:{cropSize.width}px;height:{cropSize.height}px;"
             on:mousedown|stopPropagation={startDrag}
             on:wheel|preventDefault|stopPropagation={(e) => zoom(e.deltaY > 0 ? -0.1 : 0.1)}>
            <img bind:this={imageElement} src={imageUrl} alt="crop source" class="src"
                 style="transform:translate({pos.x}px,{pos.y}px) scale({scale});transform-origin:0 0;"
                 on:load={fit} draggable="false" />
            {#if doCrop && showGrid}
                <div class="grid">{#each Array(9) as _}<i></i>{/each}</div>
            {/if}
        </div>

        {#if showCoords && dims}
            <div class="readout">
                {#if conforming && !doCrop}Original: {dims.w}×{dims.h}px • added as-is (no re-encode)
                {:else}
                    {#if doScale}Output: {dims.w}×{dims.h}px{:else}Size: {dims.w}×{dims.h}px (source){/if}
                    {#if opts?.convert} → {opts.convert.toUpperCase()}{/if}
                    {#if locked} • aspect locked{/if}
                {/if}
            </div>
        {/if}

        {#if doCrop}
            <div class="zoom">
                <button type="button" on:click|preventDefault={() => zoom(-0.2)}>−</button>
                <span>{Math.round(scale * 100)}%</span>
                <button type="button" on:click|preventDefault={() => zoom(0.2)}>+</button>
                <button type="button" class="reset" on:click|preventDefault={fit}>Reset</button>
            </div>
        {/if}

        {#if error}<div class="err">⚠️ {error}</div>{/if}

        <div class="actions">
            {#if showCancelAll}
                <button type="button" class="ghost" on:click|preventDefault={cancelAll} disabled={processing}>Skip remaining</button>
            {/if}
            <button type="button" class="secondary" on:click|preventDefault={cancel} disabled={processing}>{cancelLabel}</button>
            <button type="button" class="primary" on:click|preventDefault={confirm} disabled={processing}>
                {processing ? 'Processing…'
                    : (confirmLabel || (doCrop ? 'Apply crop' : (conforming ? 'Add image as-is' : 'Apply')))}
            </button>
        </div>
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
    h3 { margin: 0 0 4px; }
    .queue-pos { margin: 0 0 4px; color: #1c7fc7; font-size: .8rem; font-weight: bold; }
    .hint { margin: 0 0 14px; color: #666; font-size: .85rem; }
    .crop-toggle { display: inline-flex; align-items: center; gap: 6px; margin: 0 0 12px; font-size: .9rem; cursor: pointer; }
    .crop-toggle input { cursor: pointer; }
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
    .batch-warning { background: #fff8e1; border: 1px solid #f0e0a0; border-radius: 4px; color: #7a6000; padding: 10px; margin-bottom: 12px; font-size: .85rem; text-align: left; overflow-wrap: anywhere; }
    .batch-warning ul { padding-left: 20px; margin: 6px 0; max-height: 100px; overflow-y: auto; }
    .batch-warning p { margin: 6px 0 0; }
    .err { color: darkred; margin: 8px 0; font-size: .85rem; }
    .actions { display: flex; gap: 10px; margin-top: 14px; }
    .actions button { flex: 1; padding: 10px; border: none; border-radius: 6px; font-weight: bold; cursor: pointer; }
    .actions .primary { background: #1c7fc7; color: #fff; }
    .actions .secondary { background: #e7e7e7; }
    .actions .ghost { background: transparent; color: darkred; border: 1px solid #e0c0c0; font-weight: normal; flex: 0 0 auto; padding: 10px 12px; }
</style>
