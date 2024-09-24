<script>
    import AddContent from "../add_content.svelte";
    import allContent from '../../../generated/content.js';

    export let showAdd, showEditor, env;
    let title = "All Content";
    let filteredContent = allContent;
    let active = "listing";
</script>

<div class="plenti-content plenti-modal" on:click|stopPropagation>
    <div class="plenti-column-1">
        <div class="plenti-type-reset">
            <button 
                class="plenti-type-reset-selector"
                on:click={() => {
                    title = "All Content";
                    filteredContent = allContent;
                }}
            >
                All
            </button>
        </div>
        <div class="plenti-types">
            <div>Types:</div>
            {#each env?.types as content_type}
                <div class="plenti-type">
                    <button 
                        class="plenti-type-selector"
                        on:click={() => {
                            title = content_type;
                            filteredContent = allContent.filter(c => c?.type == content_type);
                        }}
                    >
                        {content_type}
                    </button>
                </div>
            {/each}
        </div>
        <div class="plenti-single-types">
            <div>Single Types:</div>
            {#each env?.singleTypes as single_type}
                <div class="plenti-single-type">
                    <button 
                        class="plenti-single-type-selector"
                        on:click={() => {
                            title = single_type;
                            filteredContent = allContent.filter(c => c?.type == single_type);
                        }}
                    >
                        {single_type}
                    </button>
                </div>
            {/each}
        </div>
    </div>
    <div class="plenti-column-2">
        <div class="plenti-selectors">
            <div class="plenti-selector {active === 'listing' ? 'active' : ''}" on:click={() => active = 'listing'}>
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-upload" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2 -2v-2" />
                <polyline points="7 9 12 4 17 9" />
                <line x1="12" y1="4" x2="12" y2="16" />
                </svg>
                <span>Listing</span>
            </div>
            <div class="plenti-selector {active === 'add' ? 'active' : ''}" on:click={() => active = 'add'}>
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-circle-plus" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <circle cx="12" cy="12" r="9" />
                <line x1="9" y1="12" x2="15" y2="12" />
                <line x1="12" y1="9" x2="12" y2="15" />
                </svg>
                <span>Add New</span>
            </div>
        </div>
        {#if active === 'listing'}
            <h2>{title}:</h2>
            <div class="plenti-content-items">
                {#each filteredContent as c}
                    <a href="{c?.path}" class="plenti-content-item">
                        <div class="plenti-content-item-filename">{c?.filename}</div>
                        <div class="plenti-content-item-path">
                            <svg  xmlns="http://www.w3.org/2000/svg"  width="18"  height="18"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-link"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 15l6 -6" /><path d="M11 6l.463 -.536a5 5 0 0 1 7.071 7.072l-.534 .464" /><path d="M13 18l-.397 .534a5.068 5.068 0 0 1 -7.127 0a4.972 4.972 0 0 1 0 -7.071l.524 -.463" /></svg>
                            <span>{c?.path}</span>
                        </div>
                    </a>
                {/each}
            </div>
        {/if}
        {#if active === 'add'}
            <AddContent bind:showAdd bind:showEditor {env} />
        {/if}
    </div>
</div>

<style>
    .plenti-content.plenti-modal {
        flex-direction: row;
        gap: 40px;
    }
    .plenti-column-1 {
        flex-basis: 250px;
        border-right: 1px solid gainsboro;
        padding-right: 40px;
    }
    .plenti-column-2 {
        flex-grow: 1;
    }
    button {
        width: 100%;
        padding: 5px;
        background: none;
        box-shadow: none;
        border: 1px solid gainsboro;
        border-radius: 5px;
        margin: 6px 0;
    }
    .plenti-content-items {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 10px;
        overflow-y: scroll;
        max-height: 100%;
    }
    .plenti-content-item {
        text-decoration: none;
        color: black;
        padding: 20px;
        border-radius: 5px;
        border: 1px solid gainsboro;
    }
    .plenti-content-item-path {
        font-size: small;
        display: flex;
        align-content: center;
        gap: 4px;
    }
</style>