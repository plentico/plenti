<script>
    import MediaFilters from './media_filters.svelte';
    import MediaGrid from './media_grid.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';

    export let media, changingMedia, showMediaModal;
    let filters = [];
    let enabledFilters = [];
    let selectedMedia = [];
    $: filteredMedia = [];

    const downloadFiles = () => {
        selectedMedia.forEach(mediaFile => {
            const a = document.createElement('a');
            a.setAttribute( 'href', mediaFile );
            a.setAttribute( 'download', mediaFile.substring(mediaFile.lastIndexOf('/')+1) );
            a.click();
        });
    }

    // Create objects that can be used by GitLab API
    $: mediaList = selectedMedia.map(i => {
        return {file: i, contents: null};
    });

    const removeMedia = () => {
        selectedMedia.forEach(m => {
            media = media.filter(i => i != m);
        });
        selectedMedia = [];
    }
</script>

<div class="media-wrapper">
    <MediaFilters bind:media bind:filters bind:enabledFilters bind:filteredMedia bind:changingMedia />
    <MediaGrid files={filteredMedia} bind:selectedMedia={selectedMedia} bind:changingMedia bind:showMediaModal />
</div>
{#if selectedMedia.length > 0} 
    <ButtonWrapper>
        <Button
            on:click={downloadFiles}
            buttonText="Download selected"
        />
        <Button
            afterSubmit={removeMedia}
            bind:commitList={mediaList}
            buttonText="Delete Selected Media"
            buttonStyle="secondary"
            action="delete"
            encoding="text"
        />
    </ButtonWrapper>
{/if}

<style>
    .media-wrapper {
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }
</style>