<script>
    import { onDestroy, tick } from 'svelte';
    import { isImagePath, isDocPath } from '../media_checker.js';
    import { parseImageOptions, transformImage, centerCoverSelection } from '../crop-engine.js';
    import { conformsToImageOptions } from '../crop-engine.js';
    import { pendingMedia } from '../pending_media.js';

    export let field, showMediaModal, changingMedia, localMediaList;
    export let uploadContext;   // set to a field context in swapMedia(); admin_menu owns/resets it
    export let schema = null, parentKeys = null;

    // Resolve THIS field's schema config: the top-level key, else the parent
    // object key (one level of nesting). Read-only — never mutates siblings.
    function resolveFieldKey(schema, parentKeys) {
        if (!schema || !parentKeys) return null;
        if (schema[parentKeys]) return parentKeys;
        // Component schemas are keyed relative to the component's `fields`
        // object, but the form path includes it ("fields.image.src") — accept
        // both conventions so an unprefixed _schema.json key still applies.
        const unwrapped = parentKeys.replace(/^fields\./, '');
        if (schema[unwrapped]) return unwrapped;
        const dot = parentKeys.lastIndexOf('.');
        if (dot > -1 && schema[parentKeys.slice(0, dot)]) return parentKeys.slice(0, dot);
        return null;
    }
    // null when the field has no image options -> behaves as an ordinary field.
    $: imageOptions = parseImageOptions(schema, resolveFieldKey(schema, parentKeys));

    // A media value is a string path OR an object { src, alt, ... }. fieldSrc is
    // the path; setFieldSrc writes it back format-preservingly (keeps alt etc.).
    $: fieldSrc = typeof field === 'string' ? field : (field?.src ?? '');
    function setFieldSrc(newSrc) {
        field = (field && typeof field === 'object') ? { ...field, src: newSrc } : newSrc;
    }
    // Schema specs are applied silently on selection (machine default) — there
    // is no field-level format control; review lives in the upload modal.
    $: displaySrc = ($pendingMedia, pendingMedia.previewUrl(fieldSrc)) || fieldSrc;

    let cropError = '';
    let processing = false;

    function loadImage(src) {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = () => resolve(img);
            img.onerror = () => reject(new Error('Could not load the source image.'));
            img.src = src;
        });
    }

    // #364 core: enforce the field's schema when a NEW image path enters the field
    // — a Library-tab pick (via this changingMedia reactive) or a fresh upload
    // (via onSavedPath after the gateway stages it). Both deliver a canonical
    // path; a fresh upload is still deferred until the page save.
    // changingMedia is SHARED across every media field, so a pick must be
    // claimed by exactly the field that opened the picker. The claim is
    // uploadContext IDENTITY: swapMedia() puts this instance's own context
    // object into uploadContext when the picker opens — long before any pick,
    // so no ordering race with the pick's modal close — and admin_menu resets
    // it on an abandoned close (the X) or a standalone open. Without this, a
    // field whose picker was closed without picking still passed the
    // field===originalMedia guard and captured the NEXT pick made for a
    // DIFFERENT field (its crop modal opened on top of the real one).
    let lastHandled;
    let selectionRequest = 0;
    let destroyed = false;
    onDestroy(() => { destroyed = true; selectionRequest++; });
    $: if (changingMedia && uploadContext === myPickContext && field === originalMedia
            && changingMedia !== fieldSrc && changingMedia !== lastHandled) {
        lastHandled = changingMedia;
        handleNewSelection(changingMedia);
    }
    async function handleNewSelection(newPath) {
        if (destroyed) return;
        const request = ++selectionRequest;
        processing = false;
        cropError = '';
        const recrop = pendingMedia.sourceOf(newPath) ?? newPath;
        // A just-staged (deferred) asset isn't served from media/ yet — LOAD it
        // from its in-memory preview; the PATH stays the persisted identity.
        const loadUrl = pendingMedia.previewUrl(newPath) ?? newPath;
        // Explicit "keep the original" from the upload review modal (Reset +
        // confirm): the field's schema must NOT be re-applied to this raw
        // passthrough. One-shot flag on this instance's own pick context.
        const bypassSchema = !!myPickContext?.bypassSchema;
        if (myPickContext) myPickContext.bypassSchema = false;
        if (!imageOptions || !isImagePath(newPath) || bypassSchema) {
            // A Library pick runs inside a reactive flush, after fieldSrc was
            // computed. Resume in the next flush so indirect field assignment
            // also invalidates the thumbnail (direct uploads run outside it).
            await tick();
            if (destroyed || request !== selectionRequest) return;
            setFieldSrc(newPath);                  // ordinary field / non-image
            return;
        }
        let probe;
        try {
            probe = await loadImage(loadUrl);
        } catch (error) {
            if (request !== selectionRequest) return;
            cropError = error instanceof Error ? error.message : 'The selected image could not be loaded.';
            return;
        }
        if (request !== selectionRequest) return;
        // Maintainer-confirmed short-circuit (#364): an asset that already meets
        // this field's spec is referenced directly — no derivative copy.
        if (conformsToImageOptions(
            { width: probe.naturalWidth, height: probe.naturalHeight, path: newPath },
            imageOptions,
        )) {
            setFieldSrc(newPath);
            return;
        }
        // Silent machine default: centered cover crop to the configured aspect
        // (when crop is enabled) + scale/convert — NO forced modal. The upload
        // review modal pre-applies the same specs for optional adjustment,
        // re-cropping from the staged ORIGINAL (recrop) at full fidelity.
        optimizeToField(loadUrl, newPath, recrop, null, probe, request);
    }

    // Silent automatic processing (machine default), no modal. Crop-enabled
    // specs use a centered cover selection matching the configured aspect;
    // crop:false specs contain-scale. Keeps the previous field value untouched
    // if the transform fails. `preloaded` skips a second decode when the caller
    // already probed the image (conformance check).
    async function optimizeToField(loadUrl, namePath, recrop, objectUrl, preloaded, request) {
        if (processing) return;
        const previous = field;
        processing = true;
        cropError = '';
        try {
            const image = preloaded ?? await loadImage(loadUrl);
            if (request !== selectionRequest) return;
            const selection = imageOptions.crop !== false
                ? centerCoverSelection(image, imageOptions)
                : null;
            const result = await transformImage(image, selection, imageOptions, namePath);
            if (request !== selectionRequest) return;
            pendingMedia.add(result.filePath, result.blob, recrop ?? result.filePath);
            setFieldSrc(result.filePath);
        } catch (error) {
            if (request !== selectionRequest) return;
            field = previous;
            cropError = error instanceof Error ? error.message : 'The image could not be processed.';
        } finally {
            if (objectUrl) URL.revokeObjectURL(objectUrl);
            if (request === selectionRequest) processing = false;
        }
    }

    // A field-launched upload was staged by the Media gateway, which handed
    // back its canonical path (via admin_menu's finishFieldUpload).
    // Feed that path into the SAME selection logic a Library-tab pick uses, so the
    // field's schema processing (crop/optimize) still runs. This handler OWNS its
    // errors: the canonical asset remains staged, and the previous field value
    // is retained on processing failure. Nothing is persisted until page save.
    // onSavedPath is the existing handoff name; it does not imply persistence.
    async function onSavedPath(path) {
        try {
            await handleNewSelection(path);
        } catch (error) {
            cropError = error instanceof Error ? error.message : 'The selected image could not be applied to this field.';
        }
    }

    // --- media-swap entry point ---
    let originalMedia;
    let myPickContext = null;   // this instance's claim on the open picker
    const swapMedia = () => {
        // Opening another picker abandons any selection still being decoded.
        selectionRequest++;
        processing = false;
        lastHandled = undefined;
        originalMedia = field;
        // Open the Media picker with this field's context. A fresh upload flows
        // through the gateway and comes back via onSavedPath; a Library pick
        // flows through the changingMedia reactive, gated on this exact context
        // object still being the current uploadContext (see the claim note
        // above). Either way the field only ever receives a path (never a File).
        myPickContext = { kind: 'field', onSavedPath, imageOptions };
        uploadContext = myPickContext;
        changingMedia = fieldSrc;
        showMediaModal = true;
    }

    // If an img path is 404, load the data image instead
    const loadDataImage = imgEl => {
        let src = imgEl.target.attributes.src.nodeValue;
        let allImg = document.querySelectorAll('img[src="' + src + '"]');
        allImg.forEach(i => {
            localMediaList.forEach(mediaItem => {
                if(mediaItem.file === fieldSrc) {
                    i.src = mediaItem.contents;
                }
            });
        });
    }
</script>

<div class="thumbnail-wrapper">
    {#if isImagePath(fieldSrc)}
        <img src="{displaySrc}" alt="click to change thumbnail" class="thumbnail" on:error={imgEl => loadDataImage(imgEl)} />
    {:else if isDocPath(fieldSrc)}
        <embed src="{displaySrc}" class="thumbnail" />
    {/if}
    <!-- Hover action: Change Media. Schema specs (when the field has them) are
         applied SILENTLY on selection; review/adjust happens in the upload
         review modal — no field-level format button. -->
    <div class="field-actions">
        <button class="swap" on:click|preventDefault={swapMedia}>Change Media</button>
    </div>
    {#if processing}
        <div class="processing">Optimizing…</div>
    {/if}
</div>
{#if cropError}
    <div class="field-error">⚠️ {cropError}</div>
{/if}

<style>
    .thumbnail-wrapper {
        height: 115px;
        overflow: hidden;
        position: relative;
    }
    .thumbnail {
        max-width: 200px;
    }
    .field-actions {
        /* Prototype geometry: the hover actions span the full field row (two
           ~200px buttons side by side), not just the 200px thumbnail — keeps
           "Change Media" / "Edit crop" legible at 1rem. */
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        height: 115px;
        display: flex;
        opacity: 0;
        transition: opacity .15s;
    }
    .thumbnail-wrapper:hover .field-actions {
        opacity: 1;
    }
    .field-actions button {
        flex: 1 1 50%;
        border: 0;
        cursor: pointer;
        color: white;
        font-size: 1rem;
        font-weight: bold;
    }
    .field-actions .swap {
        background-color: rgba(0, 0, 0, .75);
    }
    .field-actions .crop {
        background-color: #1c7fc7;
    }
    .field-actions .crop:hover {
        background-color: #15679f;
    }
    .processing {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(0, 0, 0, .6);
        color: white;
        font-weight: bold;
    }
    .field-error {
        /* The edit tray inherits the site's (often dark) background — darkred
           was unreadable there, which made a failed selection look like a
           silent no-op. Needs to stay legible on light AND dark trays. */
        color: #e05252;
        background: rgba(255, 255, 255, .9);
        border-radius: 3px;
        padding: 4px 6px;
        font-size: .85rem;
        margin-top: 4px;
    }
</style>
