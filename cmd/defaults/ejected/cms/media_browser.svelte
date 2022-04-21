<script>
    import { getAssets } from './get_assets.js';
    import { env } from '../env.js';

    let assetsDir = env.baseurl ? 'assets/' : '/assets/';
    let links = new Promise(() => {});
    let allFiles = [];
    let filters = [];
    const readDir = async (dir) => {
        links = await getAssets(dir);
        links.forEach(link => {
            let linkPath = dir + link.innerHTML;
            if (linkPath.includes('.')) {
                allFiles = [...allFiles, linkPath];
            } else {
                let filter = link.innerHTML;
                filter = filter.endsWith('/') ? filter.slice(0, -1) : filter;
                filters = [...filters, filter];
                readDir(linkPath);
            }
        });
    }
    readDir(assetsDir);

    let enabledFilters = [];
    const toggleFilter = filter => {
        console.log('fire');
        if (!filter) {
            enabledFilters = [];
        } else {
            if (enabledFilters.includes(filter)) {
                enabledFilters = enabledFilters.filter(f => f !== filter);
            } else {
                enabledFilters = [...enabledFilters, filter];
            }
        }
        allFiles = allFiles; // Force #each loop in template to rerender
    }
    const applyFilters = allFiles => {
        if (enabledFilters.length > 0) {
            let fileList = [];
            enabledFilters.forEach(filter => {
                fileList = [...fileList, ...allFiles.filter(linkPath => {
                    let parts = linkPath.split("/");
                    return parts.includes(filter) && !fileList.includes(linkPath);
                })];
            });
            return fileList;
        } else {
            return allFiles;
        }
    }
</script>

<div class="media-wrapper">
    {#await links}
        Loading...    
    {:then _}
        <div class="filters">
            {#each filters as filter}
                <div on:click={toggleFilter(filter)} class="filter{enabledFilters.includes(filter) ? ' active' : ''}">{filter}</div>
            {/each}
            {#if enabledFilters.length > 0}
                <div on:click={() => toggleFilter(false)} class="close">
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="20" height="20" viewBox="5 5 14 14" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                </div>
            {/if}
        </div>
        <div class="media-browser">
        {#each applyFilters(allFiles) as link}
            <div class="media">
                {#if link.endsWith('.pdf')}
                    <embed src="{link}" type="application/pdf">
                {:else}
                    <img src={link} />
                {/if}
            </div>
        {/each}
        </div>
    {/await}
</div>

<style>
    .media-wrapper {
        padding: 20px;
    }
    .media-browser {
        display: flex;
        flex-wrap: wrap;
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
        max-width: 200px;
    }
    .filters {
        padding: 6px 10px;
        margin-bottom: 10px;
        display: flex;
        gap: 10px;
        background: gainsboro;
        border-radius: 5px;
        align-items: center;
    }
    .filter {
        border-radius: 5px;
        display: inline-block;
        padding: 2px 10px;
        border: 1px solid;
        cursor: pointer;
    }
    .filter.active {
        background-color: gray;
        color: white;
    }
    .close {
        cursor: pointer;
        margin-left: auto;
        display: flex;
    }
</style>