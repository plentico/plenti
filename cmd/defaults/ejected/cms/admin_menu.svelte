<script>
    export let user, content;
    import JSONEditor from "./json_editor.svelte";
    import MediaBrowser from "./media_browser.svelte";
    import FileUpload from "./file_upload.svelte";
    import VisualEditor from "./visual_editor.svelte";

    let showMedia = false;
    const toggleMedia = () => {
        showMedia = !showMedia;
    }
    let activeMedia = "upload";
    const setActiveMedia = selected => {
      activeMedia = selected;
    }

    let showEditor = false;
    const toggleEditor = () => {
        showEditor = !showEditor;
    }
    let activeEditor = "visual";
    const setActiveEditor = selected => {
      activeEditor = selected;
    }
</script>

<div class="spacer"></div>
<nav>
  <a href="." id="home">
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-home-2" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <polyline points="5 12 3 12 12 3 21 12 19 12" />
      <path d="M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-7" />
    </svg>
    Home
  </a>
  <a href="." on:click|preventDefault={toggleEditor}>
    {#if showEditor}
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-eye" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <circle cx="12" cy="12" r="2" />
      <path d="M22 12c-2.667 4.667 -6 7 -10 7s-7.333 -2.333 -10 -7c2.667 -4.667 6 -7 10 -7s7.333 2.333 10 7" />
    </svg>
    View
    {:else}
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-pencil" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <path d="M4 20h4l10.5 -10.5a1.5 1.5 0 0 0 -4 -4l-10.5 10.5v4" />
      <line x1="13.5" y1="6.5" x2="17.5" y2="10.5" />
    </svg>
    Edit
    {/if}
  </a>
  <span class="gap"></span>
  <a href=".">
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-circle-plus" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <circle cx="12" cy="12" r="9" />
      <line x1="9" y1="12" x2="15" y2="12" />
      <line x1="12" y1="9" x2="12" y2="15" />
    </svg>
    Add
  </a>
  <a href="." on:click|preventDefault={toggleMedia}>
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-photo" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <line x1="15" y1="8" x2="15.01" y2="8" />
      <rect x="4" y="4" width="16" height="16" rx="3" />
      <path d="M4 15l4 -4a3 5 0 0 1 3 0l5 5" />
      <path d="M14 14l1 -1a3 5 0 0 1 3 0l2 2" />
    </svg>
    Media
  </a>
  <a href="." on:click|preventDefault={$user.logout}>
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-logout" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <path d="M14 8v-2a2 2 0 0 0 -2 -2h-7a2 2 0 0 0 -2 2v12a2 2 0 0 0 2 2h7a2 2 0 0 0 2 -2v-2" />
      <path d="M7 12h14l-3 -3m0 6l3 -3" />
    </svg>
    Logout
  </a>
</nav>

{#if showEditor}
  <div class="sidenav">
    <div class="selectors">
      <div class="selector {activeEditor === 'visual' ? 'active' : ''}" on:click={() => setActiveEditor('visual')}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-table" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <rect x="4" y="4" width="16" height="16" rx="2" />
          <line x1="4" y1="10" x2="20" y2="10" />
          <line x1="10" y1="4" x2="10" y2="20" />
        </svg>
        <span>Visual</span>
      </div>
      <div class="selector {activeEditor === 'code' ? 'active' : ''}" on:click={() => setActiveEditor('code')}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-braces" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M7 4a2 2 0 0 0 -2 2v3a2 3 0 0 1 -2 3a2 3 0 0 1 2 3v3a2 2 0 0 0 2 2" />
          <path d="M17 4a2 2 0 0 1 2 2v3a2 3 0 0 0 2 3a2 3 0 0 0 -2 3v3a2 2 0 0 1 -2 2" />
        </svg>
        <span>Code</span>
      </div>
    </div>
    {#if activeEditor === 'code'}
      <JSONEditor bind:content={content} />
    {:else}
      <VisualEditor bind:content={content} />
    {/if}
  </div>
{/if}

{#if showMedia}
  <div class="modal-wrapper" on:click={toggleMedia}>
    <div class="modal" on:click|stopPropagation>
      <div class="selectors">
        <div class="selector {activeMedia === 'upload' ? 'active' : ''}" on:click={() => setActiveMedia('upload')}>
          <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-upload" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
            <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
            <path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2 -2v-2" />
            <polyline points="7 9 12 4 17 9" />
            <line x1="12" y1="4" x2="12" y2="16" />
          </svg>
          <span>Upload</span>
        </div>
        <div class="selector {activeMedia === 'library' ? 'active' : ''}" on:click={() => setActiveMedia('library')}>
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
        <MediaBrowser />
      {:else}
        <FileUpload />
      {/if}
    </div>
  </div>
{/if}

<style>
    .spacer {
        padding: 20px;
    }
    nav {
        background-color: white;
        display: flex;
        box-shadow: 0px 1px 2px rgb(207 207 207);
        position: fixed;
        width: 100%;
        z-index: 10;
        top: 0;
    }
    svg {
        margin-right: 5px;
    }
    a {
      align-items: center;
      display: flex;
      text-decoration: none;
      color: black;
      flex-grow: 0;
      margin: 5px 10px;
    }
    a:last-of-type {
      flex: initial;
    }
    .gap {
      flex-grow: 1;
    }
    .sidenav {
      height: calc(100% - 40px);
      width: 500px;
      position: fixed;
      z-index: 1;
      top: 0;
      left: 0;
      background-color: whitesmoke;
      overflow-x: hidden;
      padding-top: 40px;
      transition: 0.5s;
      box-shadow: 1px 0px 2px rgb(207 207 207);
    }
    .sidenav + :global(main) {
      margin-left: 500px;
    }
    .selectors {
      display: flex;
    }
    .selector {
      display: flex;
      flex-grow: 1;
      flex-basis: 0;
      align-items: center;
      justify-content: center;
      padding: 5px;
      border: 1px solid gainsboro;
      background-color: whitesmoke;
      cursor: pointer;
    }
    .selector.active {
      background-color: white;
    }
    .modal-wrapper {
      z-index: 99999;
      position: fixed;
      inset: 0px;
      display: flex;
      -webkit-box-pack: center;
      justify-content: center;
      -webkit-box-align: center;
      align-items: center;
      transition: background-color 0.2s ease 0s, opacity 0.2s ease 0s;
      background-color: rgba(0, 0, 0, 0.6);
    }
    .modal {
      box-shadow: rgb(68 74 87 / 15%) 0px 4px 12px 0px, rgb(68 74 87 / 25%) 0px 1px 3px 0px;
      background-color: rgb(255, 255, 255);
      border-radius: 5px;
      height: 80%;
      width: 80%;
      max-width: 1200px;
      padding: 20px;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
</style>