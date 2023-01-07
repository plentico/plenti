<script>
    import { isMediaPath } from './media_checker.js';
    export let media, changingMedia;
    export let filters = [];
    export let enabledFilters = [];
    export let filteredMedia = []; 
    export let singleSelect = false;

    const mediaPathToArray = media => {
        // Create an array of path segments.
        let allFolders = media.split('/')
        // Get the index right after the media folder (works with and without baseurl).
        let cut = allFolders.findIndex(i => i === "media") + 1;
        // Remove "media" folder and last (filename) elements.
        return allFolders.slice(cut, -1);
    }

    const parentMediaIndex = (media, filters) => {
        // Return position of filter in filters array,
        // if it's a subpath of current media file (if not found return -1)
        return filters.findIndex(filter => media.join('').includes(filter.join('')));
    }

    for (const mediaFile of media) {
        if (isMediaPath(mediaFile)) {
            // Turn media path into array of subfolders
            let mediaFolders = mediaPathToArray(mediaFile); 
            // Make sure we're not adding empty filters
            if (mediaFolders.length > 0) {
                // Get the index of any parent folders that have already been added
                let subfolderIndex = parentMediaIndex(mediaFolders, filters);
                // Check if a parent folder was found
                if (subfolderIndex === -1) {
                    // No previously added filter is a subpath of this media path,
                    // so add the media file path to filters
                    filters = [...filters, mediaFolders];
                } else {
                    // Parent path has already been added,
                    // replace with more complete path containing child folders
                    filters[subfolderIndex] = mediaFolders;
                }
            }
        }
    }

    const mediaMatchesFilter = (mediaFile, filters) => {
        // Compare arrays in exact order by converting to strings
        return filters.find(filter => mediaFile.join('') === filter.join(''));
    }

    // Filter media
    $: filteredMedia = media.filter(mediaFile => {
        // Show all media if no filter is applied, or
        // Show specific media file if it's in the enabled filters
        return !enabledFilters.length || mediaMatchesFilter(mediaPathToArray(mediaFile), enabledFilters);
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

    if (changingMedia) {
        // Apply filters from current media when swapping for new media
        toggleFilter(mediaPathToArray(changingMedia));
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