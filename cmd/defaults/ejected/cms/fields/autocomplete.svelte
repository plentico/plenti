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
                    if (content.fields[field].includes(input)) {
                        results = [...results, content[option.result]];
                    }
                });
            });
        });
    }
</script>

<input bind:value={input} on:keydown={search} />

{#if results}
    <ul>
        {#each results as result}
            <li>{result}</li>
        {/each}
    </ul>
{/if}