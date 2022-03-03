<script>
    export let field, label, content;
    export let key = false;
</script>

{#if field === null}
    <div>{field} is null</div>
{:else if field === undefined}
    <div>{field} is undefined</div>
{:else if field.constructor === "".constructor}
    {#if field.length < 50}
        {#if key !== false}
            <input id="{label}" type="text" bind:value={content.fields[label][key]} />
        {:else}
            <input id="{label}" type="text" bind:value={content.fields[label]} />
        {/if}
    {:else}
        {#if key !== false}
            <textarea id="{label}" rows="5" bind:value={content.fields[label][key]}></textarea>
        {:else}
            <textarea id="{label}" rows="5" bind:value={content.fields[label]}></textarea>
        {/if}
    {/if}
{:else if field.constructor === true.constructor}
    <input id="{label}" type="checkbox" bind:checked={content.fields[label]} /><span>{field}</span>
{:else if field.constructor === [].constructor}
    <fieldset>
        <legend>{label}</legend>
        {#each field as value, key}
            <svelte:self field={value} {label} bind:content={content} {key} />
        {/each}
    </fieldset>
{:else if field.constructor === ({}).constructor}
    {#each Object.entries(field) as [key, value]}
        <div>
            <label>{key}</label>
            <svelte:self field={value} {label} bind:content={content} {key} />
        </div>
    {/each}
{/if}

<style>
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
    }
</style>