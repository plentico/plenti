<script>
    import allContent from '../../../generated/content.js';
    export let schema, parentKeys, field;

    let input, results, loading, option;
    console.log("reference: " + parentKeys)

    let deepCloneContent = structuredClone(allContent);

    const search = () => {
        loading = true;
        results = [];
        schema[parentKeys].options.forEach(option => {
            let filteredContent = deepCloneContent.filter(c => c.type === option.type);
            filteredContent.forEach(content => {
                option.search.forEach(field => {
                    if (content.fields.hasOwnProperty(field) && content.fields[field].includes(input)) {
                        let parts = option.result.split(".");
                        let newResult = parts[0] === "*" ? content : content[parts[0]];
                        if (parts.length > 1) {
                            parts.slice(1).forEach(part => {
                                newResult = newResult[part]
                            });
                        }
                        results = [...results, newResult];
                    }
                });
            });
        });
        if (input === "") {
            results = [];
        }
        setTimeout(() => {
            loading = false;
        }, "500")
    }
    const makeSelection = () => {
        results = [];
        field = option;
        input = "";
    }
    const removeTag = tag => {
        field = "";
    }
</script>

<div class="reference">
    <div class="input-wrapper">
        <div class="tags">
            {#if field !== ""}
                <span class="tag">
                    {field}
                    <svg on:click={() => removeTag(field)} xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="15" height="15" viewBox="0 0 24 24" stroke-width="2" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                </span>
            {/if}
        </div>
        <input bind:value={input} on:keyup={search} />
    </div>
    <div class="load-icon">
        {#if loading}
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-loader-2" width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="gray" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                <path d="M12 3a9 9 0 1 0 9 9"></path>
            </svg>
        {:else}    
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-circle" width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="gray" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                <circle cx="12" cy="12" r="9"></circle>
            </svg>
        {/if}
    </div>
    {#if results && results.length > 0}
        <select bind:value={option} size={results.length === 1 ? 2 : results.length}>
            {#each results as result}
                <option on:click={makeSelection}>{result}</option>
            {/each}
        </select>
    {/if}
</div>

<style>
    .reference {
        position: relative;
    }
    .input-wrapper {
        background: white;
        border: 1px solid gainsboro;
        overflow-y: hidden;
        height: 37px;
        display: flex;
    }
    input {
        border: none;
        outline: none;
    }
    select {
        position: absolute;
        max-height: 200px;
        top: 37px;
        left: 0;
        width: 100%;
        z-index: 1;
    }
    option {
        padding: 7px;
    }
    .tags {
        display: flex;
        gap: 7px;
        padding: 7px;
        align-items: center;
    }
    .tag {
        font-family: sans-serif;
        font-size: small;
        white-space: nowrap;
        background-color: gainsboro;
        display: flex;
        gap: 5px;
        padding: 2px 5px;
        border-radius: 4px;
    }
    .tag svg {
        cursor: pointer;
    }
    .load-icon {
        position: absolute;
        right: 6px;
        top: 8px;
    }
    .icon-tabler-loader-2 {
        -webkit-animation: spin .5s linear infinite;
        -moz-animation: spin .5s linear infinite;
        animation: spin .5s linear infinite;
    }
    @-moz-keyframes spin { 
        100% { -moz-transform: rotate(360deg); } 
    }
    @-webkit-keyframes spin { 
        100% { -webkit-transform: rotate(360deg); } 
    }
    @keyframes spin { 
        100% { 
            -webkit-transform: rotate(360deg); 
            transform:rotate(360deg); 
        } 
    }
</style>