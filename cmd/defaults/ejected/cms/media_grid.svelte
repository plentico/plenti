<script>
    import { onMount } from 'svelte';
    import { isImage, isDoc } from './asset_checker.js';
    export let files, changingAsset, showMedia;
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
        if (changingAsset !== "") {
            changingAsset = file;
            showMedia = false;
        }
        if (selectedMedia.includes(file)) {
            selectedMedia = selectedMedia.filter(m => m !== file);
        } else {
            selectedMedia = [...selectedMedia, file];
        }
    } 
</script>

<div class="media-grid">
    {#each files as file}
        <div class="media{selectedMedia.includes(file) ? ' selected' : ''}" on:click={selectMedia(file)}>
            {#if isDoc(file)}
                <embed src="{file}" type="application/pdf" />
            {:else if isImage(file)}
                <img src={file} />
            {/if}
            <div class="filename">{file.split("/").pop()}</div>
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
</style>