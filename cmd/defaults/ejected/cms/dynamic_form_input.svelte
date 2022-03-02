<script>
    export let field, label, content;
</script>

{#if field === null}
    <div>{field} is null</div>
{:else if field === undefined}
    <div>{field} is undefined</div>
{:else if field.constructor === "".constructor}
    {#if field.length < 50}
        <input id="{label}" type="text" bind:value={field} />    
    {:else}
        <textarea id="{label}" rows="5" bind:value={field}></textarea>
    {/if}
{:else if field.constructor === true.constructor}
    <input id="{label}" type="checkbox" bind:checked={field} /><span>{field}</span>
{:else if field.constructor === [].constructor}
    {#each field as value, i}
        <svelte:self field={value} />
    {/each}
{:else if field.constructor === ({}).constructor}
    {#each Object.entries(field) as [key, value]}
        <div>
            <label>{key}</label>
            <svelte:self field={value} />
        </div>
    {/each}
{/if}