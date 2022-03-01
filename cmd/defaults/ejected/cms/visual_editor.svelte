<script>
  export let content;
</script>

<form>
  {#each Object.entries(content.fields) as [label, field]}
    <label>{label}</label>
    {#if field === null}
        <div>{field} is null</div>
    {:else if field === undefined}
        <div>{field} is undefined</div>
    {:else if field.constructor === "".constructor}
      <input bind:value={content.fields[label]} />
    {:else if field.constructor === [].constructor}
      {#each field as value, i}
        <input bind:value={content.fields[label][i]} />
      {/each}
    {:else if field.constructor === ({}).constructor}
      {#each Object.keys(field) as key}
        <input bind:value={content.fields[label][key]} />
      {/each}
    {/if}
  {/each}
</form>