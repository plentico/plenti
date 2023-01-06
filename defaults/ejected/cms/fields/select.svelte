<script>
    export let schema, parentKeys, field;
</script>

<select bind:value={field}>
    {#each schema[parentKeys].options as option}
        <option>{option}</option>
    {/each}
</select>