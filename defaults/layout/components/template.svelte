<script>
  export let type;

  let path;
  let copyText = "Copy";
  const copy = async () => {
    if (!navigator.clipboard) {
      return
    }
    try {
      copyText = "Copied";
      await navigator.clipboard.writeText(path.innerHTML);
      setTimeout(() => copyText = "Copy", 500);
    } catch (err) {
      console.error('Failed to copy!', err)
    }
  }
</script>

<div class="template">
  <span>Template:</span>
  <pre>
    <code bind:this={path} class="{copyText}">layout/content/{type}.svelte</code>
    <button on:click={copy}>{copyText}</button>
  </pre>
</div>

<style>
  .template {
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
  code.copied {
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