<script>
    import { getAssets } from './get_assets.js';
    import { env } from '../env.js';

    let assetsDir = env.baseurl ? 'assets/' : '/assets/';
    let [links, images] = [new Promise(() => {}), new Promise(() => {})];
    let allFiles = [];
    const readDir = async (dir) => {
        ({ links, images } = await getAssets(dir));
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
            <img src={link} />
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
    img {
        max-width: 200px;
    }
</style>