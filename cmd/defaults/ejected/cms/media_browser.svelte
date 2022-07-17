<script>
    import MediaFilters from './media_filters.svelte';
    import MediaGrid from './media_grid.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';

    export let assets, changingAsset, showMedia;
    let filters = [];
    let enabledFilters = [];
    let selectedMedia = [];

    // Filter assets
    $: filteredAssets = assets
        .filter(asset => 
            // Show all if no filters selected
            enabledFilters.length == 0 ||
            // Or make sure the folder is in the filepath for the asset
            asset
                .split('/')
                // Remove first (assets folder) and last (filename) elements.
                .slice(1, -1)
                .some(folder => enabledFilters.includes(folder))
        );

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

    const removeAssets = () => {
        selectedMedia.forEach(m => {
            assets = assets.filter(i => i != m);
        });
        selectedMedia = [];
    }
</script>

<div class="media-wrapper">
    <MediaFilters bind:assets bind:filters bind:enabledFilters />
    <MediaGrid files={filteredAssets} bind:selectedMedia={selectedMedia} bind:changingAsset bind:showMedia />
</div>
{#if selectedMedia.length > 0} 
    <ButtonWrapper>
        <button on:click={downloadFiles}>Download selected</button> 
        <div class="delete-wrapper" on:click={removeAssets}>
            <Button bind:commitList={mediaList} buttonText="Delete Selected Media" action="delete" encoding="text" />
        </div>
    </ButtonWrapper>
{/if}

<style>
    .media-wrapper {
        margin: 20px 0;
        overflow: hidden;
    }
</style>