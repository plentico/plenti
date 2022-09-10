<script>
    export let schema, parentKeys, field;
</script>

{#each schema[parentKeys].options as option}
    <div class="radio">
        <label>
            <input
                id="{option}" 
                type="radio"
                name="radio"
                checked={field === option}
                on:click={() => {
                    field = "";
                    field = option;
                }}
            /> {option}</label>
    </div>
{/each}