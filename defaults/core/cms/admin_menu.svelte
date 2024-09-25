<script>
    export let user, content, shadowContent;
    import ModalWrapper from "./modals/modal_wrapper.svelte";
    import MediaModal from "./modals/media_modal.svelte";
    import ContentModal from "./modals/content_modal.svelte";
    import EditTray from "./edit_tray.svelte";
    import allMedia from '../../generated/media.js';
    import { env } from '../../generated/env.js';

    let mediaPrefix = env.baseurl ? '' : '/';
    let media = allMedia.map(media => mediaPrefix + media);

    let showContentModal = false;
    let showMediaModal = false;
    const toggleMediaModal = () => {
        showMediaModal = !showMediaModal;
        changingMedia = "";
    }

    let showEditor = false;
    const toggleEditor = () => {
        showEditor = !showEditor;
    }

    const horizontalSlide = () => {
      return 	{
        delay: 0,
        duration: 100,
        css: t =>
          'width: ' + t * 500 + 'px;'
      };
    }

    let changingMedia = "";
    let localMediaList = [];
</script>

<div class="spacer"></div>
<nav class="admin-menu">
  <a href="{env.baseurl ? '.' : '/'}" class="home">
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-home-2" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <polyline points="5 12 3 12 12 3 21 12 19 12" />
      <path d="M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-7" />
    </svg>
    Home
  </a>
  <a href="." class="{showEditor ? 'view' : 'edit'}" on:click|preventDefault={toggleEditor}>
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
  <a href="." class="content" on:click|preventDefault={() => showContentModal = !showContentModal}>
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icons-tabler-outline icon-tabler-file-text" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <path d="M14 3v4a1 1 0 0 0 1 1h4" />
      <path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z" />
      <path d="M9 9l1 0" />
      <path d="M9 13l6 0" />
      <path d="M9 17l6 0" />
    </svg>
    Content
  </a>
  <a href="." class="media" on:click|preventDefault={toggleMediaModal}>
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-photo" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <line x1="15" y1="8" x2="15.01" y2="8" />
      <rect x="4" y="4" width="16" height="16" rx="3" />
      <path d="M4 15l4 -4a3 5 0 0 1 3 0l5 5" />
      <path d="M14 14l1 -1a3 5 0 0 1 3 0l2 2" />
    </svg>
    Media
  </a>
  <a href="." class="logout" on:click|preventDefault={$user.logout}>
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-logout" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
      <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
      <path d="M14 8v-2a2 2 0 0 0 -2 -2h-7a2 2 0 0 0 -2 2v12a2 2 0 0 0 2 2h7a2 2 0 0 0 2 -2v-2" />
      <path d="M7 12h14l-3 -3m0 6l3 -3" />
    </svg>
    Logout
  </a>
</nav>

{#if showContentModal}
  <ModalWrapper on:click={() => showContentModal = !showContentModal}>
    <ContentModal bind:showContentModal bind:showEditor {env} />
  </ModalWrapper>
{/if}

{#if showMediaModal}
  <ModalWrapper on:click={toggleMediaModal}>
    <MediaModal 
      bind:media
      bind:changingMedia
      bind:showMediaModal
      bind:localMediaList
      {mediaPrefix}
      {user}
    />
  </ModalWrapper>
{/if}

<div class={showEditor ? "sidenav-wrapper" : ""}>
{#if showEditor}
  <div transition:horizontalSlide|local class={showEditor ? "sidenav" : ""}>
    <EditTray 
      bind:content
      bind:showMediaModal
      bind:changingMedia
      bind:localMediaList
      bind:shadowContent
      {user}
    />
  </div>
{/if}
</div>

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
      height: 100%;
      width: 500px;
      position: fixed;
      z-index: 1;
      top: 0;
      left: 0;
      overflow-x: hidden;
      padding-top: 40px;
      transition: 0.5s;
      box-shadow: 1px 0px 2px rgb(207 207 207);
      box-sizing: border-box;
    }
    :global(body > div),
    :global(body > section),
    :global(body > main) {
      transition: margin-left .1s ease-in-out;
    }
    .sidenav-wrapper + :global(div),
    .sidenav-wrapper + :global(section),
    .sidenav-wrapper + :global(main) {
      margin-left: 500px;
    }
    :global(.plenti-selectors) {
      display: flex;
    }
    :global(.plenti-selector) {
      display: flex;
      flex-grow: 1;
      flex-basis: 0;
      align-items: center;
      justify-content: center;
      padding: 5px;
      border: 1px solid gainsboro;
      background-color: #ebebeb;
      cursor: pointer;
    }
    :global(.plenti-selector.active) {
      background-color: white;
    }
</style>
