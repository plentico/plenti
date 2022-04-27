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
        </div>
    {/each}
</div>

<style>
    .media-browser {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 10px;
        margin-bottom: 20px;
    }
    .media {
        width: 200px;
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
    }
    img, embed {
        min-width: 200px;
        min-height: 150px;
        object-fit: cover;
    }
</style>