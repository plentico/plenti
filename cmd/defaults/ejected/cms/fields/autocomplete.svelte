<script>
    import allContent from '../../content.svelte';
    export let schema, parentKeys, field;

    let input, results, loading, option;

    //input = field;
    if (field.constructor !== [].constructor) {
        field = [field];
    }

    const search = () => {
        loading = true;
        results = [];
        schema[parentKeys].options.forEach(option => {
            let filteredContent = allContent.filter(c => c.type = option.type);
            filteredContent.forEach(content => {
                option.search.forEach(field => {
                    if (content.fields.hasOwnProperty(field)
                        && content.fields[field].includes(input)) {
                        results = [...results, content[option.result]];
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
        field = [...field, option]
        //input = field;
    }
    const removeTag = tag => {
        field = field.filter(t => t !== tag);
    }
</script>

<div class="autocomplete">
    <input bind:value={input} on:keyup={search} />
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
        <select bind:value={option} size={results.length}>
            {#each results as result}
                <option on:click={makeSelection}>{result}</option>
            {/each}
        </select>
    {/if}
    {#if field.constructor === [].constructor}
        <div class="tags">
            {#each field as tag}
                <span class="tag">
                    {tag}
                    <svg on:click={() => removeTag(tag)} xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                </span>
            {/each} 
        </div>
    {/if}
</div>

<style>
    .autocomplete {
        position: relative;
    }
    input {
        width: 100%;
        box-sizing: border-box;
        height: 37px;
        padding: 7px;
        padding-right: 32px;
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
        gap: 5px;
        padding: 10px 0;
    }
    .tag {
        background-color: gainsboro;
        display: flex;
        gap: 5px;
        padding: 2px 5px;
        border-radius: 4px;
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