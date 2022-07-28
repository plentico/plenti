<script>
    import { isAssetPath } from './assets_checker.js';
    export let assets, changingAsset;
    export let filters = [];
    export let enabledFilters = [];
    export let filteredAssets = []; 
    export let singleSelect = false;

    const assetPathToArray = asset => {
        // Create an array of path segments.
        let allFolders = asset.split('/')
        // Get the index right after the assets folder (works with and without baseurl).
        let cut = allFolders.findIndex(i => i === "assets") + 1;
        // Remove "assets" folder and last (filename) elements.
        return allFolders.slice(cut, -1);
    }

    const parentAssetIndex = (asset, filters) => {
        // Return position of filter in filters array,
        // if it's a subpath of current asset (if not found return -1)
        return filters.findIndex(filter => asset.join('').includes(filter.join('')));
    }

    for (const asset of assets) {
        if (isAssetPath(asset)) {
            // Turn asset path into array of subfolders
            let assetFolders = assetPathToArray(asset); 
            // Make sure we're not adding empty filters
            if (assetFolders.length > 0) {
                // Get the index of any parent folders that have already been added
                let subfolderIndex = parentAssetIndex(assetFolders, filters);
                // Check if a parent folder was found
                if (subfolderIndex === -1) {
                    // No previously added filter is a subpath of this asset path,
                    // so add the asset path to filters
                    filters = [...filters, assetFolders];
                } else {
                    // Parent path has already been added,
                    // replace with more complete path containing child folders
                    filters[subfolderIndex] = assetFolders;
                }
            }
        }
    }

    const assetMatchesFilter = (asset, filters) => {
        // Compare arrays in exact order by converting to strings
        return filters.find(filter => asset.join('') === filter.join(''));
    }

    // Filter assets
    $: filteredAssets = assets.filter(asset => {
        // Show all assets if no filter is applied, or
        // Show specific asset if it's in the enabled filters
        return !enabledFilters.length || assetMatchesFilter(assetPathToArray(asset), enabledFilters);
    });

    const getFilterSubGroup = (filter, filterGroup) => {
        // Get filters position inside full group array
        let filterIndex = filterGroup.findIndex(f => f === filter);
        // Get array from first folder to where this filter was found (cut off nested folders after that)
        let filterSubGroup = filterGroup.slice(0, filterIndex + 1);
        // Return array of filter and parent folders (no children folders)
        return filterSubGroup;
    }

    const filterIsEnabled = (enabledFilters, filterSubGroup) => {
        // Compare exact order of arrays by converting to strings
        return enabledFilters.find(f => f.join('') === filterSubGroup.join(''));
    }

    const toggleFilter = (filterSubGroup) => {
        if (singleSelect) {
            clearFilters();
        }
        if (filterIsEnabled(enabledFilters, filterSubGroup)) {
            // Remove filter
            enabledFilters = enabledFilters.filter(current => current.join('') !== filterSubGroup.join(''));
        } else {
            // Add filter and force update for enabled filters
            enabledFilters = [...enabledFilters, filterSubGroup];
        }
    }

    const clearFilters = () => {
        enabledFilters = [];
    }

    if (changingAsset) {
        // Apply filters from current asset when swapping for new asset
        toggleFilter(assetPathToArray(changingAsset));
    }

</script>

<div class="filters-wrapper">
    <div class="filters">
    {#each filters as filterGroup}
        <div class="filter-group">
            {#each filterGroup as filter}
                <button on:click={() => toggleFilter(getFilterSubGroup(filter, filterGroup))} class="filter{filterIsEnabled(enabledFilters, getFilterSubGroup(filter, filterGroup)) ? ' active' : ''}">{filter}</button>
            {/each}
        </div>
    {/each}
    </div>
    {#if enabledFilters.length > 0}
        <button on:click={clearFilters} class="close">
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="20" height="20" viewBox="5 5 14 14" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
        </button>
    {/if}
</div>

<style>
    .filters-wrapper {
        display: flex;
        margin-top: 20px;
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
        font-weight: bold;
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        font-size: .8rem;
    }
    .filter {
        display: inline-block;
        padding: 4px 10px;
        color: #1c7fc7;
        font-weight: bold;
    }
    .filter.active {
        background-color: #1c7fc7;
        color: white;
    }
    .close {
        padding: 3px 0;
        margin-left: auto;
        display: flex;
    }
    button {
        background-color: transparent;
        border: none;
        cursor: pointer;
    }
</style>