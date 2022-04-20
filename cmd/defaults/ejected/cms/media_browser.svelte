<script>
    import { getAssets } from './get_assets.js';
    import { env } from '../env.js';

    let assetsDir = env.baseurl ? 'assets/' : '/assets/';
    let links = new Promise(() => {});
    let allFiles = [];
    const readDir = async (dir) => {
        links = await getAssets(dir);
        links.forEach(link => {
            let linkPath = dir + link.innerHTML;
            if (linkPath.includes('.')) {
                allFiles = [...allFiles, linkPath];
            } else {
                readDir(linkPath);
            }
        });
    }
    readDir(assetsDir);
</script>

<div class="media-browser">
{#await links}
    Loading...    
{:then _}
    {#each allFiles as link}
        <div class="media">
            {#if link.endsWith('.pdf')}
                <embed src="{link}" type="application/pdf">
            {:else}
                <img src={link} />
            {/if}
        </div>
    {/each}  
{/await}
</div>

<style>
    .media-browser {
        padding: 20px;
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
</style>