<script>
    import MediaGrid from './media_grid.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';
    import { isAsset } from './assets_checker.js';

    export let assets, changingAsset, showMedia;
    let filters = [];
    let enabledFilters = [];
    let selectedMedia = [];

    for (const asset of assets) {
        if (isAsset(asset)) {
            // Create an array of path segments.
            let allFolders = asset.split('/')
            // Get the index right after the assets folder (works with and without baseurl).
            let cut = allFolders.findIndex(i => i === "assets") + 1;
            // Remove first (assets folder) and last (filename) elements.
            let folders = allFolders.slice(cut, -1);
            if (folders.length > 0 && !filters.includes(folders)) {
                // Get the index of any parent folders that have already been added
                let subfolderIndex = filters.findIndex(val => {
                    let filterStr = val.join('');
                    let folderStr = folders.join('');
                    return folderStr.includes(filterStr);
                });
                // Check if a parent folder was found
                if (subfolderIndex === -1) {
                    // No subpaths match this path, so add it
                    filters = [...filters, folders];
                } else {
                    // Parent path has already been added,
                    // replace with more complete path containing child folders
                    filters[subfolderIndex] = folders;
                }
            }
        }
    }

    const toggleFilter = filter => {
        if (enabledFilters.includes(filter)) {
            // Remove filter
            enabledFilters = enabledFilters.filter(current => current != filter);
        } else {
            // Add filter and force update for enabled filters
            enabledFilters = [...enabledFilters, filter];
        }
    }

    const clearFilters = () => {
        enabledFilters = [];
    }

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
    <div class="filters-wrapper">
        <div class="filters">
        {#each filters as filterGroup}
            <div class="filter-group">
                {#each filterGroup as filter}
                    <div on:click={() => toggleFilter(filter)} class="filter{enabledFilters.includes(filter) ? ' active' : ''}">{filter}</div>
                {/each}
            </div>
        {/each}
        </div>
        {#if enabledFilters.length > 0}
            <div on:click={clearFilters} class="close">
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="20" height="20" viewBox="5 5 14 14" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
            </div>
        {/if}
    </div>
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
    .filters-wrapper {
        display: flex;
    }
    .filters {
        display: flex;
        gap: 10px;
        border-radius: 5px;
        align-items: center;
        flex-wrap: wrap;
    }
    .filter-group {
        border-radius: 6px;
        cursor: pointer;
        font-weight: bold;
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        font-size: .8rem;
    }
    .filter {
        display: inline-block;
        padding: 4px 10px;
    }
    .filter.active {
        background-color: #1c7fc7;
        color: white;
    }
    .close {
        cursor: pointer;
        padding: 5px 0;
        margin-left: auto;
        display: flex;
    }
</style>