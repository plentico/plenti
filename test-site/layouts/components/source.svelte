<script>
  export let content, source;

  let templateEl;
  let contentEl;
  let copied;
  const copy = async (el) => {
    if (!navigator.clipboard) {
      return
    } 
    try {
      await navigator.clipboard.writeText(el.innerHTML);
      copied = el;
      setTimeout(() => copied = null, 500);
    } catch (err) {
      console.error('Failed to copy!', err)
    }
  }
</script>

{#if source.layout}
  <div>
    <span>Layout:</span>
    <pre>
      <code bind:this={templateEl} class:selected="{copied === templateEl}">layouts/content/{content.type}.svelte</code>
      <button on:click={() => copy(templateEl)}>{copied === templateEl ? 'copied' : 'copy'}</button>
    </pre>
  </div>  
{/if}

{#if source.content}
  <div>
    <span>Content:</span>
    <pre>
      <code bind:this={contentEl} class:selected="{copied === contentEl}">content/{content.type === 'index' ? '' : content.type + '/'}{content.filename}</code>
      <button on:click={() => copy(contentEl)}>{copied === contentEl ? 'copied' : 'copy'}</button>
    </pre>
  </div>
{/if}

<style>
  div {
    display: flex;
    align-items: center;
  }
  pre {
    display: flex;
    padding-left: 5px;
  }
  code {
      background-color: var(--base);
      padding: 5px 10px;
  }
  code.selected {
      background-color: var(--accent);
  }
  button {
    border: 1px solid rgba(0,0,0,.1);
    background: white;
    padding: 4px;
    border-top-right-radius: 5px;
    border-bottom-right-radius: 5px;
    cursor: pointer;
  }
</style>