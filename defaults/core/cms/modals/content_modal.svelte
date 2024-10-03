<script>
    import AddContent from "../add_content.svelte";
    import allContent from '../../../generated/content.js';

    export let showContentModal, showEditor, env;
    let selectedType = "";
    let filteredContent = allContent;
    let showAdd = false;

    function truncate(str, maxLength) {
        return str.length > maxLength ? str.slice(0, maxLength) + '...' : str;
    }
    function removeExtension(str, ext) {
        const fullExt = ext.startsWith('.') ? ext : '.' + ext;
        return str.endsWith(fullExt) ? str.slice(0, -fullExt.length) : str;
    }
</script>

<div class="plenti-content plenti-modal" on:click|stopPropagation>
    <div class="plenti-column-1">
        <div class="plenti-type-reset">
            <button 
                class="plenti-type-reset-selector {selectedType == '' ? 'selected' : ''}"
                on:click={() => {
                    selectedType = "";
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
                        class="plenti-type-selector {selectedType == content_type ? 'selected' : ''}"
                        on:click={() => {
                            selectedType = content_type;
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
                        class="plenti-single-type-selector {selectedType == single_type ? 'selected' : ''}"
                        on:click={() => {
                            selectedType = single_type;
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
            <div class="plenti-selector {showAdd ? '' : 'active'}" on:click={() => showAdd = false}>
                <svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="#2c3e50" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon icon-tabler icons-tabler-outline icon-tabler-list">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                    <path d="M9 6l11 0" />
                    <path d="M9 12l11 0" />
                    <path d="M9 18l11 0" />
                    <path d="M5 6l0 .01" />
                    <path d="M5 12l0 .01" />
                    <path d="M5 18l0 .01" />
                </svg>
                <span>Listing</span>
            </div>
            <div class="plenti-selector {showAdd ? 'active' : ''}" on:click={() => showAdd = true}>
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-circle-plus" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                    <circle cx="12" cy="12" r="9" />
                    <line x1="9" y1="12" x2="15" y2="12" />
                    <line x1="12" y1="9" x2="12" y2="15" />
                </svg>
                <span>Add New</span>
            </div>
        </div>
        {#if showAdd && !env?.singleTypes?.includes(selectedType)}
            <AddContent bind:showContentModal bind:showAdd bind:showEditor bind:selectedType {env} />
        {:else}
            <div class="plenti-content-items">
                <div class="plenti-content-items-grid">
                    {#if !env?.singleTypes?.includes(selectedType)}
                        <button
                            class="add-new" 
                            on:click={() => showAdd = true}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-circle-plus" width="30" height="30" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                                <circle cx="12" cy="12" r="9" />
                                <line x1="9" y1="12" x2="15" y2="12" />
                                <line x1="12" y1="9" x2="12" y2="15" />
                            </svg>
                            Add {selectedType}
                        </button>
                    {/if}
                    {#each filteredContent as c}
                        <a href="{c?.path}" class="plenti-content-item">
                            <div class="plenti-content-item-filename">{truncate(removeExtension(c?.filename, ".json"), 28)}</div>
                            <div class="plenti-content-item-path">
                                <svg  xmlns="http://www.w3.org/2000/svg"  width="18"  height="18"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-link"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 15l6 -6" /><path d="M11 6l.463 -.536a5 5 0 0 1 7.071 7.072l-.534 .464" /><path d="M13 18l-.397 .534a5.068 5.068 0 0 1 -7.127 0a4.972 4.972 0 0 1 0 -7.071l.524 -.463" /></svg>
                                <span>{truncate(c?.path, 30)}</span>
                            </div>
                        </a>
                    {/each}
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .plenti-content.plenti-modal {
        flex-direction: row;
        gap: 40px;
    }
    .plenti-column-1 {
        flex-basis: 0;
        flex-grow: 1;
        border-right: 1px solid gainsboro;
        padding-right: 40px;
    }
    .plenti-column-2 {
        flex-basis: 0;
        flex-grow: 7;
    }
    button {
        width: 100%;
        padding: 5px;
        background: none;
        box-shadow: none;
        border: 1px solid gainsboro;
        border-radius: 5px;
        margin: 6px 0;
        cursor: pointer;
    }
    button.selected {
        background-color: gainsboro;
    }
    button.add-new {
        margin: 0;
        font-size: initial;
        border: 1px dashed;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 5px;
        min-height: 80px;
    }
    button.add-new:hover {
        background-color: gainsboro;
    }
    .plenti-content-items {
        overflow-y: auto;
        max-height: calc(100% - 63px);
        margin: 20px 0;
    }
    .plenti-content-items-grid {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 10px;
        word-break: break-all;
        grid-auto-rows: 1fr;
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
    svg.icon-tabler-link {
        min-width: 18px;
        min-height: 18px;
    }
</style>