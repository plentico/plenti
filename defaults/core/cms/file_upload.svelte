<script>
    import MediaFilters from './media_filters.svelte';
    import MediaGrid from './media_grid.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';
    // NOTE: keep every import line under ~80 chars. Svelte's printer wraps a
    // longer one across lines in the compiled output, and Plenti's regex
    // import-rewriter (cmd/build/compile.go) mis-spans a file containing MORE
    // THAN ONE wrapped import: its greedy multi-line branch swallows everything
    // between the first `import {` and the last line-start `} from`, corrupting
    // the SSR component ("missing ) after argument list"). The generated
    // svelte/internal import already wraps once this component uses {#each},
    // so every hand-written import here must stay single-line.
    import { createEventDispatcher } from 'svelte';
    import { onMount } from 'svelte';
    import { onDestroy } from 'svelte';
    import ImageCropModal from './fields/image_crop_modal.svelte';
    import { transformImage, blobToDataURL } from './crop-engine.js';
    import { LIBRARY_OPTIMISE_DEFAULTS } from './crop-engine.js';
    import { conformsToImageOptions } from './crop-engine.js';
    import { libraryFingerprint, libraryOutputPath } from './library_optimise.js';
    import { pendingMedia } from './pending_media.js';
    import { STANDALONE_UPLOAD_CONTEXT } from './upload_context.js';
    import { classifyFile } from './upload_queue.js';

    export let media, changingMedia, showMediaModal, localMediaList, mediaPrefix, user;
    const dispatch = createEventDispatcher();   // 'saved' (filePath) on a field-launched save
    // Explicit upload context — NEVER inferred from changingMedia (a field with no
    // current image has changingMedia === '', which would wrongly read as standalone).
    //   { kind: 'standalone' }                         — top-nav Media library
    //   { kind: 'field', onSavedPath(path) { … } }     — a field's "Change Media"
    export let uploadContext = STANDALONE_UPLOAD_CONTEXT;
    // The standalone upload SESSION (queue + keyed payload store) is OWNED by
    // media_modal so it survives this view's remounts (tab switches). All
    // mutations go through sessionOps — the owner's self-assignment is the
    // single reactive root, so everything derived from `session` re-renders.
    export let session = null;
    export let sessionOps = null;
    $: isFieldUpload = uploadContext?.kind === 'field';   // REACTIVE, not a cached const
    let enabledFilters = [];

    // ── derived session state (recomputes on every owner notify) ─────────────
    $: queue = session?.queue ?? null;
    // Fail-closed Save derivation: resolved items mapped through the keyed
    // store; ANY missing payload blocks Save instead of silently committing.
    $: resolvedPayloads = queue
        ? queue.items.filter(i => i.state === 'resolved').map(i => session.staged.get(i.id))
        : null;
    $: canSave = queue
        ? (queue.canSave && resolvedPayloads.length > 0 && resolvedPayloads.every(Boolean))
        : localMediaList.length > 0;   // queue-null holds only approved payloads (the owner's teardown guarantees it)
    $: unresolvedCount = queue ? queue.unresolvedCount : 0;
    $: failedItems = queue ? queue.failedItems() : [];
    $: pendingReviewCount = unresolvedCount - failedItems.length;
    $: remainingSkippable = queue ? queue.skippableCount : 0;
    // "Image X of Y" over image-type items only (passthrough is not modal-driven).
    $: queuePosition = (() => {
        if (!queue || !currentItem || currentItem.type !== 'image') return null;
        const images = queue.items.filter(i => i.type === 'image');
        const index = images.indexOf(currentItem) + 1;
        return index > 0 ? { index, total: images.length } : null;
    })();

    // ── the optimise-gateway modal driver ─────────────────────────────────────
    let showCropModal = false;
    // Field mode: a component-owned object URL (created + revoked here).
    // Standalone: an alias of the queue-owned item.objectUrl — never revoked here.
    let cropSourceUrl = '';
    let cropSourceFile = null; // the File being optimised (for the fingerprint)
    let sourcePath = '';      // media/<name> the OUTPUT path is derived from
    let cropError = '';
    let fieldNote = '';       // field multi-drop notice (shown via the modal's error slot)
    let processing = false;
    let currentItem = null;   // the queue item this view is presenting
    let destroyed = false;
    let mounted = false;
    let fieldUploadRequest = 0;

    // FIELD mode only — standalone URLs belong to the queue's lifecycle.
    function revokeCropUrl() {
        if (cropSourceUrl) { URL.revokeObjectURL(cropSourceUrl); cropSourceUrl = null; }
    }

    // ── ingestion conformance (#364, owner-confirmed) ─────────────────────────
    // A source that ALREADY meets the library defaults (target format, within
    // the max edge) must never be silently re-encoded: a deliberately
    // pre-optimised asset (e.g. an aggressively compressed 15 KB WebP) can come
    // out LARGER from the canvas. Such files are offered/added AS-IS — original
    // bytes, ORIGINAL name (the raw-passthrough naming contract, not the hashed
    // derivative identity), action 'create'. Gitea/GitLab reject same-name
    // creates; the existing local dev endpoint still overwrites them.
    function loadProbeImage(src) {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = () => resolve(img);
            img.onerror = () => reject(new Error('Could not load the image.'));
            img.src = src;
        });
    }
    function conformsToLibraryDefaults(probe, name) {
        return conformsToImageOptions(
            { width: probe.naturalWidth, height: probe.naturalHeight, path: name },
            LIBRARY_OPTIMISE_DEFAULTS,
        );
    }

    // Transform + name a library derivative. Returns the pieces both flows need:
    // the standalone queue builds a TRANSPORT (data URL) for its staged batch;
    // the field flow stages the BLOB in pendingMedia (deferred to the page save).
    async function deriveLibraryAsset({ image, selection, overrides, file, sourcePath }) {
        const options = { ...LIBRARY_OPTIMISE_DEFAULTS, ...(overrides || {}) };
        const result = await transformImage(image, selection, options, sourcePath);
        const fingerprint = await libraryFingerprint({ file, sourceRect: result.sourceRect, options });
        const filePath = libraryOutputPath({
            sourcePath, fingerprint,
            width: result.width, height: result.height, mime: result.actualMime,
        });
        return { filePath, blob: result.blob };
    }
    async function buildDerivativeItem(args) {
        const { filePath, blob } = await deriveLibraryAsset(args);
        // A data URL in `contents` (the provider strips the prefix); derivative → upsert.
        return { action: 'upsert', encoding: 'base64', file: filePath, contents: await blobToDataURL(blob) };
    }

    // ── the queue driver ──────────────────────────────────────────────────────
    // One loop per view instance. Ownership is enforced by the queue's
    // SESSION-WIDE run claims, not a component-local counter: a chain that loses
    // its claim (teardown, batch replacement, a newer claim) drops its late
    // result; an in-flight chain from a PREVIOUS mount keeps its claim and
    // completes its one item (exactly one read/transform), after which this
    // mount's reactive resume takes over the queue.
    async function driveQueue() {
        while (!destroyed) {
            const q = session?.queue;
            if (!q) return;
            const item = q.current;
            if (!item) {                                   // drained — Save owns the rest
                if (showCropModal || currentItem) { showCropModal = false; currentItem = null; }
                return;
            }
            if (q.hasActiveRun(item)) return;              // another chain is driving it
            if (item.state === 'awaiting_decision') {      // user decision pending (or remount re-bind)
                bindImageModal(item);
                return;
            }
            const token = q.claimRun(item);
            if (!token) return;
            currentItem = item;
            if (item.type === 'passthrough') {
                // No modal for passthrough: close any open one BEFORE the read so
                // the crop UI is never blank-but-actionable over an invisible item.
                if (showCropModal) { showCropModal = false; cropSourceUrl = ''; }
                if (!q.process(item)) { q.clearRun(item, token); return; }
                sessionOps.notify();
                let transport = null, readError = null;
                try {
                    const contents = await blobToDataURL(item.file);   // a File IS a Blob
                    transport = { action: 'create', encoding: 'base64', file: mediaPrefix + "media/" + item.file.name, contents };
                } catch (error) {
                    readError = error instanceof Error ? error : new Error('The file could not be read.');
                }
                // Late-completion gate: the session may have been torn down, the
                // batch replaced, or the claim stolen while we were reading.
                if (session?.queue !== q || !q.ownsRun(item, token)) return;
                if (transport) sessionOps.complete(item, token, transport);   // ATOMIC resolve+stage
                else sessionOps.fail(item, token, readError);
                if (destroyed) return;   // a newer mount's reactive resume continues
                continue;
            }
            // Image: make a preview URL and hand the decision to the user. No
            // async work is in flight while awaiting, so release the claim —
            // the confirm handler claims its own run.
            if (!q.prepare(item)) { q.clearRun(item, token); return; }
            // One-time conformance probe (flag travels on the item, so it
            // survives remounts). An unreadable image is NOT marked conforming
            // — it falls through to the modal, which surfaces the load error.
            if (item.conformsToLibrary === undefined) {
                let conforms = false;
                try {
                    conforms = conformsToLibraryDefaults(
                        await loadProbeImage(item.objectUrl), item.file.name);
                } catch (_) { /* leave false */ }
                // The claim may have been lost while probing (teardown/remount).
                if (session?.queue !== q || !q.ownsRun(item, token)) return;
                item.conformsToLibrary = conforms;
            }
            q.awaitDecision(item);
            q.clearRun(item, token);
            bindImageModal(item);
            sessionOps.notify();
            return;
        }
    }

    function bindImageModal(item) {
        if (destroyed) return;
        if (currentItem === item && showCropModal) return;   // already presenting it
        currentItem = item;
        cropError = '';
        cropSourceFile = item.file;
        cropSourceUrl = item.objectUrl;   // queue-owned lifecycle
        sourcePath = mediaPrefix + "media/" + item.file.name;
        showCropModal = true;
    }

    // Resume whenever the owner notifies (a previous mount's chain completing,
    // a new batch, a removal…). Idempotent: claimed / awaiting / drained states
    // return without side effects beyond (re)binding the modal.
    onMount(() => { mounted = true; driveQueue(); });
    // Tab switches must NOT tear the session down — media_modal owns teardown on
    // ITS destroy. Field-mode preview URLs are this component's to revoke.
    onDestroy(() => { destroyed = true; if (isFieldUpload) revokeCropUrl(); });
    $: if (mounted && session) resumeIfIdle(session);
    function resumeIfIdle(_session) {
        if (!destroyed && !processing) driveQueue();
    }

    async function onLibraryCropConfirm(event) {
        if (processing) return;
        processing = true;
        cropError = '';
        if (isFieldUpload) {
            try {
                const { filePath, blob } = await deriveLibraryAsset({
                    image: event.detail.image, selection: event.detail.selection,
                    overrides: event.detail.overrides, file: cropSourceFile, sourcePath,
                });
                // DEFERRED (maintainer-confirmed, #364): stage the canonical asset
                // in pendingMedia instead of committing now. It flushes WITH the
                // page save — canonical + any placement derivative + content land
                // in ONE commit, and an abandoned edit persists nothing. The modal
                // owner (admin_menu) closes+resets and hands the PATH to the field;
                // the field previews it via pendingMedia until it is persisted.
                pendingMedia.add(filePath, blob, filePath);
                revokeCropUrl();
                dispatch('saved', filePath);
            } catch (error) {
                cropError = error instanceof Error ? error.message : 'The image could not be processed.';
            } finally {
                processing = false;
            }
            return;
        }
        // Standalone: claim the decision run for the current item. A lost claim
        // (teardown / remount race) means the late result is dropped, not staged.
        const q = session?.queue;
        const item = currentItem;
        const token = q ? q.claimRun(item) : null;
        if (!token) { processing = false; return; }
        try {
            // Conforming item confirmed WITHOUT a crop → add the ORIGINAL BYTES
            // under the original name ('create': Gitea/GitLab reject conflicts;
            // the local dev endpoint still overwrites). Ticking Crop opts back into
            // the derivative flow — cropping inherently re-encodes.
            const addAsIs = item.conformsToLibrary === true && !event.detail.selection;
            const transport = addAsIs
                ? { action: 'create', encoding: 'base64',
                    file: mediaPrefix + "media/" + item.file.name,
                    contents: await blobToDataURL(item.file) }
                : await buildDerivativeItem({
                    image: event.detail.image, selection: event.detail.selection,
                    overrides: event.detail.overrides, file: item.file,
                    sourcePath: mediaPrefix + "media/" + item.file.name,
                });
            if (session?.queue !== q || !q.ownsRun(item, token)) { processing = false; return; }
            sessionOps.complete(item, token, transport);   // ATOMIC resolve+stage
        } catch (error) {
            // A failed image transform keeps the item awaiting (retryable in the
            // open modal) — release our claim and show the error where it happened.
            if (session?.queue === q) q.clearRun(item, token);
            cropError = error instanceof Error ? error.message : 'The image could not be processed.';
            processing = false;
            return;
        }
        processing = false;
        cropSourceUrl = '';        // the queue revoked the item's URL at resolve
        driveQueue();              // next item (or drained → the modal closes)
    }

    // Modal "Cancel": field mode closes; standalone = SKIP THIS FILE (explicit,
    // buttons-only — the backdrop is inert in queue mode). Belt-and-braces
    // processing guard: the UI is disabled during processing anyway.
    function onLibraryCropCancel() {
        if (processing) return;
        cropError = '';
        if (isFieldUpload || !queue) {
            revokeCropUrl();
            showCropModal = false;
            return;
        }
        if (currentItem) sessionOps.skipCurrent(currentItem);   // cancels + revokes + clears claim
        cropSourceUrl = '';
        showCropModal = false;
        driveQueue();
    }

    // "Skip remaining": approved items stay staged and savable; every pending
    // AND failed item is dropped (the owner rebuilds the commit list).
    function onSkipRemaining() {
        if (processing) return;
        cropError = '';
        sessionOps.skipRemaining();
        cropSourceUrl = '';
        currentItem = null;
        showCropModal = false;
    }

    // ── FIELD-LAUNCHED single-file path (no queue, no session) ────────────────
    function optimiseLibraryFile(file) {
        cropError = '';
        cropSourceFile = file;
        cropSourceUrl = URL.createObjectURL(file);
        sourcePath = mediaPrefix + "media/" + file.name;
        showCropModal = true;
    }
    // Ingestion conformance, field flavour: an already-conforming image skips
    // the optimise modal entirely and defers AS-IS through the raw-passthrough
    // path (original bytes + name, 'create'). The field's own selection logic
    // still applies its schema — a crop-configured field opens its placement
    // crop on the returned path exactly as for a Library pick.
    async function routeFieldImage(file, request) {
        const context = uploadContext;
        const url = URL.createObjectURL(file);
        let conforms = false;
        try {
            conforms = conformsToLibraryDefaults(await loadProbeImage(url), file.name);
        } catch (_) { /* unreadable → the modal surfaces the load error */ }
        URL.revokeObjectURL(url);
        // The picker can close, change owners, or accept another file while
        // decoding. A stale probe must neither stage bytes nor hand off a path.
        if (destroyed || uploadContext !== context || request !== fieldUploadRequest) return;
        if (conforms) passthroughFieldFile(file);
        else optimiseLibraryFile(file);
    }
    function passthroughFieldFile(file) {
        const filePath = mediaPrefix + "media/" + file.name;
        // DEFERRED raw passthrough (a File IS a Blob): stages in pendingMedia and
        // flushes with the page save. Gitea/GitLab reject same-name 'create';
        // the pre-existing local dev endpoint overwrites (separate backend fix).
        pendingMedia.add(filePath, file, filePath, 'create');
        dispatch('saved', filePath);
    }

    // Entry point for BOTH input-change and drag-drop. Field-launched is
    // single-file (extra dropped files are declined WITH a notice); standalone
    // (re)starts the owner-held session — the session prop update triggers this
    // view's reactive resume, which starts the drive chain.
    const selectFile = files => {
        const list = Array.from(files || []);
        if (!list.length) return;
        if (isFieldUpload) {
            const request = ++fieldUploadRequest;
            const file = list[0];
            // (string concat, not a template literal — Plenti's SSR regex pipeline
            // mishandles user template literals nested in the render output)
            fieldNote = list.length > 1 ? 'Only one file can be used here — using ' + file.name + '.' : '';
            if (classifyFile(file) === 'image') void routeFieldImage(file, request);
            else passthroughFieldFile(file);
            return;
        }
        sessionOps?.start(list);
    };

    let filePrefix = mediaPrefix + "media/";
    $: if (enabledFilters) {
        if (enabledFilters.length > 0) {
            // Convert filter array to path. NOTE: this mutates the staged
            // transport objects in place — the keyed store holds the SAME object
            // references, so both ledgers see the rewritten paths.
            let filterPath = enabledFilters[0].join('/') + "/";
            let newPrefix = mediaPrefix + "media/" + filterPath;
            localMediaList.forEach(mediaFile => {
                mediaFile.file = mediaFile.file.replace(filePrefix, newPrefix);
            });
            // Set new prefix in case filter is switched and needs to be replaced
            filePrefix = newPrefix;
        }
    }

    let drag;
    const toggleDrag = () => {
        drag = !drag;
    }
    const dropFile = ev => {
        if (!ev.dataTransfer) return;
        // Collect dropped files, then route through the SAME entry point as the
        // input (selectFile) so field/standalone branching applies identically.
        const files = [];
        if (ev.dataTransfer.items) {
            for (let i = 0; i < ev.dataTransfer.items.length; i++) {
                if (ev.dataTransfer.items[i].kind === 'file') {
                    const file = ev.dataTransfer.items[i].getAsFile();
                    if (file) files.push(file);
                }
            }
        } else if (ev.dataTransfer.files) {
            for (let i = 0; i < ev.dataTransfer.files.length; i++) files.push(ev.dataTransfer.files[i]);
        }
        if (files.length) selectFile(files);
    }

    let selectedMedia = [];
    const removeSelectedMedia = () => {
        if (queue) {
            // Route through the owner so the bijection holds: find each selected
            // transport's resolved item and remove item + payload TOGETHER.
            selectedMedia.forEach(contents => {
                const entry = queue.items.find(i => i.state === 'resolved'
                    && session.staged.get(i.id)?.contents === contents);
                if (entry) sessionOps.removeItem(entry);
            });
            selectedMedia = [];
            return;
        }
        selectedMedia.forEach(file => {
            localMediaList = localMediaList.filter(i => i.contents !== file);
        });
        selectedMedia = [];
    }

    const getThumbnails = mediaList => mediaList.map(i => i.contents);

    // UNIVERSAL fix: every saved item enters the library as its PERSISTED PATH
    // (item.file), never its transport data URL (item.contents). Dedupe by path.
    const addUploadsToLibrary = () => {
        localMediaList.forEach(item => pendingMedia.rememberSaved(item.file, item.contents));
        const savedPaths = localMediaList.map(item => item.file).filter(Boolean);
        media = [...new Set([...media, ...savedPaths])];
    }
</script>

<div class="upload-wrapper">
    {#if !showCropModal && queue && (pendingReviewCount > 0 || failedItems.length > 0)}
        <div class="queue-status">
            {#if pendingReviewCount > 0}
                <div>{pendingReviewCount} file{pendingReviewCount === 1 ? '' : 's'} still to review — finish or skip them before saving.</div>
            {/if}
            {#each failedItems as f (f.id)}
                <div class="failed-row">
                    <span class="failed-msg">⚠️ {f.file.name} — {f.error instanceof Error ? f.error.message : String(f.error || 'could not be processed')}</span>
                    <button type="button" class="remove-failed" on:click|preventDefault={() => sessionOps.removeItem(f)}>Remove</button>
                </div>
            {/each}
        </div>
    {/if}
    {#if localMediaList.length > 0}
        <MediaFilters bind:media bind:enabledFilters singleSelect={true} {changingMedia} />
        <MediaGrid files={getThumbnails(localMediaList)} bind:selectedMedia={selectedMedia} />
        <ButtonWrapper>
            <Button
                afterSubmit={() => {
                    addUploadsToLibrary();
                    sessionOps?.endAfterSave();
                    enabledFilters=[];
                    filePrefix = mediaPrefix + "media/";
                    if(changingMedia) {
                        changingMedia = localMediaList[0]?.file;
                        showMediaModal = false;
                    }
                }}
                bind:commitList={localMediaList}
                buttonText="Save Media"
                action="create"
                encoding="base64"
                disabled={!canSave}
                retainCommitListOnFailure={true}
                {user}
            />
            {#if selectedMedia.length > 0}
                <Button
                    on:click="{removeSelectedMedia}"
                    buttonText="Discard selected"
                    buttonStyle="secondary"
                />
            {:else}
                <Button
                    on:click="{() => { if (sessionOps) sessionOps.discardAll(); else localMediaList = []; }}"
                    buttonText="Discard all"
                    buttonStyle="secondary"
                />
            {/if}
        </ButtonWrapper>
    {:else}
        <div class="upload-widgets">
            <div class="drop{drag ? ' active' : ''}"
                on:dragenter={toggleDrag}
                on:dragleave={toggleDrag}
                on:drop|preventDefault={event => dropFile(event)}
                on:dragover|preventDefault
            >
                <div class="drop-icon">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-cloud-upload" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M7 18a4.6 4.4 0 0 1 0 -9a5 4.5 0 0 1 11 2h1a3.5 3.5 0 0 1 0 7h-1" />
                        <polyline points="9 15 12 12 15 15" />
                        <line x1="12" y1="12" x2="12" y2="21" />
                    </svg>
                </div>
                <div class="drop-text">Drag a file here to upload</div>
            </div>
            <div class="or">Or</div>
            <div class="choose" on:change={event => {
                const files = Array.from(event.target.files || []);
                event.target.value = '';   // same-file re-selection must fire again
                selectFile(files);
            }}>
                <label class="file">
                    <input type="file" multiple={!isFieldUpload} aria-label="File browser">
                    <span class="file-custom"></span>
                </label>
            </div>
        </div>
    {/if}
</div>

{#if showCropModal}
    {#key cropSourceUrl}
        <ImageCropModal
            imageUrl={cropSourceUrl}
            options={LIBRARY_OPTIMISE_DEFAULTS}
            libraryMode={true}
            allowCropToggle={!isFieldUpload}
            conforming={!!currentItem?.conformsToLibrary}
            confirmLabel={isFieldUpload ? 'Use optimised image'
                : (currentItem?.conformsToLibrary ? '' : 'Add optimised image')}
            queueMode={!!queue}
            queuePosition={queuePosition}
            cancelLabel={queue ? 'Skip this file' : 'Cancel'}
            showCancelAll={!!queue && remainingSkippable > 1}
            batchFailures={failedItems.map(f => ({ name: f.file.name,
                message: f.error instanceof Error ? f.error.message
                    : String(f.error || 'could not be processed') }))}
            error={cropError || fieldNote}
            {processing}
            on:confirm={onLibraryCropConfirm}
            on:cancel={onLibraryCropCancel}
            on:cancelAll={onSkipRemaining}
        />
    {/key}
{/if}

<style>
    .upload-wrapper {
        display: flex;
        flex-direction: column;
        overflow: hidden;
        height: 100%;
    }
    .queue-status {
        margin: 8px 0;
        padding: 8px 12px;
        background: #fff8e1;
        border: 1px solid #f0e0a0;
        border-radius: 4px;
        color: #7a6000;
        font-size: .85rem;
        text-align: center;
    }
    .failed-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        margin-top: 6px;
        color: darkred;
        text-align: left;
    }
    .failed-msg {
        word-break: break-word;
    }
    .remove-failed {
        flex: 0 0 auto;
        border: 1px solid #e0c0c0;
        border-radius: 4px;
        background: #fff;
        color: darkred;
        padding: 3px 10px;
        cursor: pointer;
    }
    .upload-widgets {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        box-sizing: border-box;
    }
    .drop {
        width: 100%;
        height: 40%;
        box-sizing: border-box;
        justify-content: center;
        border: 2px dashed;
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .drop.active {
        border-color: #1c7fc7;
        background-color: gainsboro;
    }
    .or {
        margin: 20px;
    }
    .file {
        position: relative;
        cursor: pointer;
    }
    .file input {
        border-radius: 50%;
    }
    .file-custom {
        position: absolute;
        top: 0;
        right: 0;
        left: 0;
        z-index: 5;
        padding: 0.5rem 1rem;
        background-color: #fff;
        border: 0.075rem solid #ddd;
        border-radius: 0.25rem;
        -webkit-user-select: none;
        -moz-user-select: none;
        -ms-user-select: none;
        user-select: none;
    }
    .file-custom:before {
        position: absolute;
        top: -0.075rem;
        right: -0.075rem;
        bottom: -0.075rem;
        content: "Browse";
        padding: 0.5rem 1rem;
        background-color: #eee;
        border: 0.075rem solid #ddd;
        border-radius: 0 0.25rem 0.25rem 0;
    }
    .file-custom:after {
        content: "Choose file...";
    }
</style>
