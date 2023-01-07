<script>
    import { isImagePath, isDocPath } from '../media_checker.js';

    export let field, showMediaModal, changingMedia, localMediaList;

    let originalMedia;
    const swapMedia = () => {
        originalMedia = field;
        changingMedia = field;
        showMediaModal = true;
    }
    $: if (changingMedia) {
        if (field === originalMedia) {
            field = changingMedia;
        }
    }

    // If an img path is 404, load the data image instead
    const loadDataImage = imgEl => {
        // Get src from img that was clicked on in visual editor
        let src = imgEl.target.attributes.src.nodeValue;
        // Load all image on the page with that source
        // TODO: Could load images not related to this field specifically
        let allImg = document.querySelectorAll('img[src="' + src + '"]');
        allImg.forEach(i => {
            localMediaList.forEach(mediaItem => {
                // Check if the field path matches a recently uploaded file in memory
                if(mediaItem.file === field) {
                    // Set the source to the data image instead of the path that can't be found
                    i.src = mediaItem.contents; 
                }
            });
        });
    }
</script>

<div class="thumbnail-wrapper">
    {#if isImagePath(field)}
        <img src="{field}" alt="click to change thumbnail" class="thumbnail" on:error={imgEl => loadDataImage(imgEl)} />
    {:else if isDocPath(field)}
        <embed src="{field}" class="thumbnail" />
    {/if}
    <button class="swap" on:click|preventDefault={swapMedia}>Change Media</button>
</div>

<style>
    .thumbnail-wrapper {
        height: 115px;
        overflow: hidden;
        position: relative;
    }
    .thumbnail {
        max-width: 200px;
    }
    button.swap {
        cursor: pointer;
        position: absolute;
        top: 0;
        left: 0;
        width: 200px;
        height: 115px;
        border: 0;
        background-color: transparent;
        color: transparent;
        font-size: 1.25rem;
        transition: all .15s;
    }
    button.swap:hover {
        background-color: rgba(0, 0, 0, .75);
        color: white;
    }
</style>