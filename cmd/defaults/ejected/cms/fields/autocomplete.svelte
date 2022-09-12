<script>
    import allContent from '../../content.svelte';
    export let schema, parentKeys, field;

    let input, results;

    const search = () => {
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
    }
</script>

<div class="autocomplete">
    <input bind:value={input} on:keyup={search} />
    {#if results && results.length > 0}
        <select bind:value={field} size={results.length}>
            {#each results as result}
                <option>{result}</option>
            {/each}
        </select>
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
</style>