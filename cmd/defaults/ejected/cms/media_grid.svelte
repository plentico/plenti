<script>
    export let files;
    export let selectedMedia = [];

    const isImage = file => {
        let extensions = ['jpg', 'jpeg', 'png', 'webp', 'gif', 'svg', 'avif', 'apng'];
        let reImage = new RegExp("^data:image\/(?:" + extensions.join("|") + ")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
        return extensions.includes(file.substr(file.lastIndexOf('.') + 1)) || reImage.test(file);
    }

    const isPDF = file => {
        let extensions = ['pdf', 'msword'];
        let rePDF = new RegExp("^data:application\/(?:" + extensions.join("|") +")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
        return extensions.includes(file.substr(file.lastIndexOf('.') + 1)) || rePDF.test(file);
    }

    const selectMedia = file => {
        if (selectedMedia.includes(file)) {
            selectedMedia = selectedMedia.filter(m => m !== file);
        } else {
            selectedMedia = [...selectedMedia, file];
        }
    } 
</script>

<div class="media-browser">
    {#each files as file}
        <div class="media{selectedMedia.includes(file) ? ' selected' : ''}" on:click={selectMedia(file)}>
            {#if isPDF(file)}
                <embed src="{file}" type="application/pdf" />
            {:else if isImage(file)}
                <img src={file} />
            {/if}
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
    .media-browser {
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