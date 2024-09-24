<script>
    import JSONEditor from "./json_editor.svelte";
    import VisualEditor from "./visual_editor.svelte";

    export let content, showMediaModal, changingMedia, localMediaList, shadowContent, user

    let activeEditor = "visual";
    const setActiveEditor = selected => {
      activeEditor = selected;
    }
</script>

<div class="plenti-edit-tray">
    <div class="plenti-selectors">
      <div class="plenti-selector {activeEditor === 'visual' ? 'active' : ''}" on:click={() => setActiveEditor('visual')}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-table" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <rect x="4" y="4" width="16" height="16" rx="2" />
          <line x1="4" y1="10" x2="20" y2="10" />
          <line x1="10" y1="4" x2="10" y2="20" />
        </svg>
        <span>Visual</span>
      </div>
      <div class="plenti-selector {activeEditor === 'code' ? 'active' : ''}" on:click={() => setActiveEditor('code')}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-braces" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M7 4a2 2 0 0 0 -2 2v3a2 3 0 0 1 -2 3a2 3 0 0 1 2 3v3a2 2 0 0 0 2 2" />
          <path d="M17 4a2 2 0 0 1 2 2v3a2 3 0 0 0 2 3a2 3 0 0 0 -2 3v3a2 2 0 0 1 -2 2" />
        </svg>
        <span>Code</span>
      </div>
    </div>
    {#if activeEditor === 'code'}
      <JSONEditor bind:content {user} />
    {:else}
      <VisualEditor 
        bind:content
        bind:showMediaModal
        bind:changingMedia
        bind:localMediaList
        bind:shadowContent
        {user}
      />
    {/if}
</div>