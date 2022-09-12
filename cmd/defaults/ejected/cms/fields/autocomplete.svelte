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

<input bind:value={input} on:keyup={search} />

{#if results}
    <ul>
        {#each results as result}
            <li on:click={() => field = result}>{result}</li>
        {/each}
    </ul>
{/if}