<script>
    import { getLinks } from './get_media.js';
    import { env } from '../env.js';

    let assetsDir = env.baseurl ? 'assets' : '/assets';
    let links = getLinks(assetsDir);
    console.log(links);
</script>

<div class="media-browser">
{#await links}
    Loading...    
{:then links} 
    {#each links as link}
        <div class="media">
            <img src={assetsDir + "/" + link.childNodes[0].data} />
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