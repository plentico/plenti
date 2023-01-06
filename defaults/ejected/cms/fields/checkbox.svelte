<script>
    export let schema, parentKeys, field;
</script>

{#each schema[parentKeys].options as option}
    <div class="checkbox">
        <label>
            <input
                id="{option}" 
                type="checkbox"
                checked={field.includes(option)}
                on:click={() => {
                    if (field.includes(option)) {
                        field = field.filter(f => f != option);
                    } else {
                        field = [...field, option];
                    }
                }}
            /> {option}</label>
    </div>
{/each}