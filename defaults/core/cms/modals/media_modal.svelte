<script>
    // NOTE: one name per svelte import line (Plenti regex rewriter constraint).
    import { createEventDispatcher } from "svelte";
    import { onDestroy } from "svelte";
    import MediaBrowser from "../media_browser.svelte";
    import FileUpload from "../file_upload.svelte";
    import { STANDALONE_UPLOAD_CONTEXT } from "../upload_context.js";
    import { createUploadQueue } from "../upload_queue.js";

    export let media, changingMedia, showMediaModal, localMediaList, mediaPrefix, user;
    // Standalone unless a field opens the picker with its own context (Slice 2).
    export let uploadContext = STANDALONE_UPLOAD_CONTEXT;

    // Relay a field-launched save up to admin_menu (which owns close+reset+notify).
    const dispatch = createEventDispatcher();

    let activeMedia = "upload";
    const setActiveMedia = selected => {
      activeMedia = selected;
    }

    // ── the standalone upload SESSION ──────────────────────────────────────
    // Owned HERE so it survives Upload<->Library tab switches (FileUpload is a
    // remountable view over it). The queue is the sole lifecycle ledger; staged
    // transports live in a keyed payload store (queue-item id -> transport) so
    // resolved items and savable payloads form a checkable bijection.
    //   session = { queue, staged: Map<itemId, transport> } | null
    let session = null;
    const notify = () => { session = session; };   // the single reactive root

    // Derive the Save Media commit list from RESOLVED queue items mapped
    // through the keyed store. localMediaList itself is owned by admin_menu
    // (bound through here) so promoted payloads survive this modal's teardown.
    function rebuildCommitList() {
        if (!session) return;
        const resolved = session.queue.items.filter(i => i.state === 'resolved');
        const payloads = resolved.map(i => session.staged.get(i.id));
        if (payloads.some(p => !p)) {
            // Bijection breach — should be impossible via complete(); fail loud
            // in the console and closed in the UI (canSave blocks on it too).
            console.error('upload session invariant violated: a resolved item has no staged payload');
        }
        localMediaList = payloads.filter(Boolean);
    }

    // Owner-controlled session operations — passed into FileUpload so mutations
    // happen at the owner and reactivity has one root (no child-side bump()).
    const sessionOps = {
        // (Re)start a batch. Any previous queue is torn down first.
        start(files) {
            if (session) session.queue.destroy();
            session = {
                queue: createUploadQueue(files, f => URL.createObjectURL(f)),
                staged: new Map(),
            };
        },
        // ATOMIC resolve+stage — one invariant-preserving operation:
        // confirm the run still owns the item, stage the payload, transition to
        // resolved; roll the payload back if the transition is refused; rebuild.
        complete(item, token, transport) {
            if (!session || !session.queue.ownsRun(item, token)) return false;
            session.staged.set(item.id, transport);
            if (!session.queue.resolve(item)) {
                session.staged.delete(item.id);
                notify();
                return false;
            }
            rebuildCommitList();
            notify();
            return true;
        },
        fail(item, token, error) {
            if (!session || !session.queue.ownsRun(item, token)) return false;
            const done = session.queue.fail(item, error);
            notify();
            return done;
        },
        skipCurrent(item) {
            if (!session) return;
            session.queue.cancelCurrent(item);
            notify();
        },
        // "Skip remaining": resolved items (and their payloads) survive; every
        // pending AND failed item is dropped. Save enables for the approved work.
        skipRemaining() {
            if (!session) return;
            session.queue.skipRemaining();
            rebuildCommitList();
            notify();
        },
        // The failed-item affordance (and any future per-item removal).
        removeItem(item) {
            if (!session) return;
            session.staged.delete(item.id);
            session.queue.remove(item);
            rebuildCommitList();
            notify();
        },
        // After a successful "Save Media" batch commit: the staged work is
        // persisted; the session ends. localMediaList is cleared by the Button's
        // own success path (resetStatus), preserving its "Changes committed" UX.
        endAfterSave() {
            if (session) session.queue.destroy();
            session = null;
        },
        // "Discard all": throw EVERYTHING away — queue, staged payloads, list.
        discardAll() {
            if (session) session.queue.destroy();
            session = null;
            localMediaList = [];
        },
        notify,
    };

    // Explicit modal-close teardown (ordered — makes the queue-null fallback in
    // FileUpload safe by construction): skip everything non-resolved (this also
    // invalidates any active run claim, so an in-flight read/transform's late
    // completion is dropped by its ownsRun() check) -> promote the resolved
    // payloads into admin_menu's surviving localMediaList -> destroy the queue
    // and store. Nothing retained is revoked: resolved items' object URLs were
    // already revoked at resolve time and their payloads are data-URL transports.
    onDestroy(() => {
        if (!session) return;
        session.queue.skipRemaining();
        rebuildCommitList();
        session.queue.destroy();
        session = null;
    });
</script>

<div class="plenti-media plenti-modal" on:click|stopPropagation>
    <div class="plenti-selectors">
      <div class="plenti-selector {activeMedia === 'upload' ? 'active' : ''}" on:click={() => setActiveMedia('upload')}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-upload" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2 -2v-2" />
          <polyline points="7 9 12 4 17 9" />
          <line x1="12" y1="4" x2="12" y2="16" />
        </svg>
        <span>Upload</span>
      </div>
      <div class="plenti-selector {activeMedia === 'library' ? 'active' : ''}" on:click={() => setActiveMedia('library')}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-layout-grid" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <rect x="4" y="4" width="6" height="6" rx="1" />
          <rect x="14" y="4" width="6" height="6" rx="1" />
          <rect x="4" y="14" width="6" height="6" rx="1" />
          <rect x="14" y="14" width="6" height="6" rx="1" />
        </svg>
        <span>Library</span>
      </div>
    </div>
    {#if activeMedia === 'library'}
      <MediaBrowser
        bind:media
        bind:changingMedia
        bind:showMediaModal
        {user}
      />
    {:else}
      <FileUpload
        bind:media
        bind:changingMedia
        bind:showMediaModal
        bind:localMediaList
        {mediaPrefix}
        {uploadContext}
        {user}
        {session}
        {sessionOps}
        on:saved={(e) => dispatch('fieldSaved', e.detail)}
      />
    {/if}
</div>

<style>
  .plenti-modal {
    flex-direction: column;
  }
</style>
