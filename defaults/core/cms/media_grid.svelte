<script>
    import { onMount } from 'svelte';
    import { isImage, isDoc } from './media_checker.js';
    export let files, changingMedia, showMediaModal;
    export let selectedMedia = [];

    onMount(async () => {
        focus();
        window.addEventListener('blur', () => {
            let embeds = document.querySelectorAll('embed');
            embeds.forEach(embed => {
                if (document.activeElement === embed) {
                    selectMedia(embed.attributes.src.nodeValue);
                }
            });
            window.parent.focus();
        });
    });

    const selectMedia = file => {
        if (changingMedia !== "") {
            changingMedia = file;
            showMediaModal = false;
        }
        if (selectedMedia.includes(file)) {
            selectedMedia = selectedMedia.filter(m => m !== file);
        } else {
            selectedMedia = [...selectedMedia, file];
        }
    } 

    let copiedStates = files.map(() => false);
    const copy = async (file, i) => {
        copiedStates[i] = true;
		await navigator.clipboard.writeText(file ?? '');
        setTimeout(() => {
            copiedStates[i] = false;
        }, 1000);
	}
</script>

<div class="media-grid">
    {#each files as file, i}
        <div class="media{selectedMedia.includes(file) ? ' selected' : ''}" on:click={selectMedia(file)}>
            {#if isDoc(file)}
                <embed src="{file}" type="application/pdf" />
            {:else if isImage(file)}
                <img src={file} />
            {/if}
            <div class="filename">{file.split("/").pop()}</div>
            <button 
                type="button"
                class="copy"
                on:click|preventDefault|stopPropagation={() => copy(file, i)}
            >
                {#if copiedStates[i]}
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon icon-tabler icons-tabler-outline icon-tabler-checked"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 12l5 5l10 -10"/></svg>
                {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon icon-tabler icons-tabler-outline icon-tabler-link"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 15l6 -6"/><path d="M11 6l.463 -.536a5 5 0 0 1 7.071 7.072l-.534 .464"/><path d="M13 18l-.397 .534a5.068 5.068 0 0 1 -7.127 0a4.972 4.972 0 0 1 0 -7.071l.524 -.463"/></svg>
                {/if}
            </button>
            {#if selectedMedia.includes(file)}
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-check" width="65" height="65" viewBox="0 0 24 24" stroke-width="2.5" stroke="#1c7fc7" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <path d="M5 12l5 5l10 -10" />
            </svg>
            {/if}
        </div>
    {/each}
</div>

<style>
    .media-grid {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr 1fr 1fr;
        gap: 10px;
        overflow-y: scroll;
        height: 100%;
        margin: 20px 0;
    }
    .media {
        width: 190px;
        height: 150px;
        overflow: hidden;
        background-color: gainsboro;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 2px solid black;
        position: relative;
        cursor: pointer;
    }
    .filename {
        position: absolute;
        width: 100%;
        bottom: -20px;
        left: 0;
        text-align: center;
        transition: all .15s;
        background: rgba(0,0,0,.5);
        color: white;
    }
    .media:hover .filename {
        bottom: 0;
    }
    .media.selected {
        border: 2px solid #1c7fc7;
        position: relative;
        background-color: black;
    }
    .media.selected img {
        opacity: 0.5;
    }
    .icon-tabler-check {
        position: absolute;
    }
    img, embed {
        min-width: 200px;
        min-height: 150px;
        object-fit: cover;
    }
    .copy {
        display: none;
        position: absolute;
        top: 2px;
        right: 2px;
        cursor: pointer;
    }
    .media:hover .copy {
        display: block;
    }
</style>