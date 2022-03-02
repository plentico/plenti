<script>
    export let content;
</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        <div class="field">
            <label for="{label}">{label}</label>
            {#if field === null}
                <div>{field} is null</div>
            {:else if field === undefined}
                <div>{field} is undefined</div>
            {:else if field.constructor === "".constructor}
                {#if field.length < 1150}
                    <input id="{label}" type="text" bind:value={content.fields[label]} />    
                {:else}
                    <textarea id="{label}" rows="5" bind:value={content.fields[label]}></textarea>    
                {/if}
            {:else if field.constructor === true.constructor}
                <input id="{label}" type="checkbox" bind:checked={content.fields[label]} /><span>{field}</span>
            {:else if field.constructor === [].constructor}
                {#each field as value, i}
                    <input id="{label}" bind:value={content.fields[label][i]} />
                {/each}
            {:else if field.constructor === ({}).constructor}
                {#each Object.keys(field) as key}
                    <input id="{label}" bind:value={content.fields[label][key]} />
                {/each}
            {/if}
        </div>
    {/each}
</form>

<style>
    form {
        padding: 20px;
    }
    label {
        display: block;
    }
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
    }
    .field {
        margin-bottom: 20px;
    }
</style>