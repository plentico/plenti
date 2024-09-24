<script>
    import MediaBrowser from "../media_browser.svelte";
    import FileUpload from "../file_upload.svelte";

    export let media, changingMedia, showMediaModal, localMediaList, mediaPrefix, user;

    let activeMedia = "upload";
    const setActiveMedia = selected => {
      activeMedia = selected;
    }
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
        {user}
      />
    {/if}
</div>

<style>
  .plenti-modal {
    flex-direction: column;
  }
</style>