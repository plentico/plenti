<script>
    import { getAssets, baseUrl } from './get_assets.js';
    import MediaGrid from './media_grid.svelte';

    let assets = [];
    let filters = [];
    let enabledFilters = [];

    async function loadIndex() {
        assets = await getAssets();
        for (const asset of assets) {
            const folders = asset
                .split('/')
                // Remove first (assets folder) and last (filename) elements.
                .slice(1, -1);
            filters.push(...folders);
        }

        // Force update for filters
        filters = filters;
    }

    function toggleFilter(filter) {
        if (enabledFilters.includes(filter)) {
            enabledFilters = enabledFilters.filter(current => current != filter);
        } else {
            enabledFilters.push(filter);

            // Force update for enabled filters
            enabledFilters = enabledFilters;
        }
    }

    function clearFilters() {
        enabledFilters = [];
    }

    // Filter assets
    let filteredAssets;
    $: filteredAssets = assets
            .filter(asset => 
                enabledFilters.length == 0 ||
                asset
                    .split('/')
                    // Remove first (assets folder) and last (filename) elements.
                    .slice(1, -1)
                    .some(folder => enabledFilters.includes(folder))
            )
            .map(asset => baseUrl + '/' + asset);
</script>

<div class="media-wrapper">
    {#await loadIndex()}
        Loading...    
    {:then _}
        <div class="filters-wrapper">
            <div class="filters">
            {#each filters as filter}
                <div on:click={() => toggleFilter(filter)} class="filter{enabledFilters.includes(filter) ? ' active' : ''}">{filter}</div>
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
        <MediaGrid files={filteredAssets} />
    {/await}
</div>

<style>
    .media-wrapper {
        padding: 20px;
    }
    .filters-wrapper {
        display: flex;
    }
    .filters {
        margin-bottom: 10px;
        display: flex;
        gap: 10px;
        border-radius: 5px;
        align-items: center;
        flex-wrap: wrap;
    }
    .filter {
        border-radius: 6px;
        display: inline-block;
        padding: 4px 10px;
        cursor: pointer;
        font-weight: bold;
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        font-size: .8rem;
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