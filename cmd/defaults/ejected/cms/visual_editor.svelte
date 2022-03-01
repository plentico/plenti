<script>
  export let content;
</script>

<form>
  {#each Object.entries(content.fields) as [label, field]}
    <div class="field">
      <label>{label}</label>
      {#if field === null}
          <div>{field} is null</div>
      {:else if field === undefined}
          <div>{field} is undefined</div>
      {:else if field.constructor === "".constructor}
        <input type="text" bind:value={content.fields[label]} />
      {:else if field.constructor === true.constructor}
        <input type="checkbox" bind:checked={content.fields[label]} /><span>{field}</span>
      {:else if field.constructor === [].constructor}
        {#each field as value, i}
          <input bind:value={content.fields[label][i]} />
        {/each}
      {:else if field.constructor === ({}).constructor}
        {#each Object.keys(field) as key}
          <input bind:value={content.fields[label][key]} />
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