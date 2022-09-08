<script>
    export let schema, parentKeys, field;
</script>

{#each schema[parentKeys].values as value}
    <div class="checkbox">
        <label>
            <input
                id="{value}" 
                type="checkbox"
                checked={field.includes(value)}
                on:click={() => {
                    if (field.includes(value)) {
                        field = field.filter(f => f != value);
                    } else {
                        field = [...field, value];
                    }
                }}
            /> {value}</label>
    </div>
{/each}