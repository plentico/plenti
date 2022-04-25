<script>
    export let files;
    const isImage = file => {
        let extensions = ['jpg', 'jpeg', 'png', 'webp', 'gif', 'svg', 'avif', 'apng'];
        let reImage = new RegExp("^data:image\/(?:" + extensions.join("|") + ")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
        return extensions.includes(file.substr(file.lastIndexOf('.') + 1)) || reImage.test(file);
    }
    const isPDF = file => {
        let rePDF = new RegExp("^data:application\/(?:pdf)(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
        return file.endsWith('.pdf') || rePDF.test(file);

    }
</script>
<div class="media-browser">
    {#each files as file}
        {#if isPDF(file)}
            <div class="media">
                <embed src="{file}" type="application/pdf">
            </div>
        {:else if isImage(file)}
            <div class="media">
                <img src={file} />
            </div>
        {/if}
    {/each}
</div>

<style>
    .media-browser {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 10px;
    }
    .media {
        width: 200px;
        height: 150px;
        overflow: hidden;
        background-color: gainsboro;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    img, embed {
        min-width: 200px;
        min-height: 150px;
        object-fit: cover;
    }
</style>