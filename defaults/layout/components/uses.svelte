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
      setTimeout(() => copyText = "Copy", 800);
    } catch (err) {
      console.error('Failed to copy!', err)
    }
  }
</script>

<details>
  <summary>Uses the "{type}" template</summary>
  <pre>
    <code bind:this={path}>layout/content/{type}.svelte</code><button on:click={copy}>{copyText}</button>
  </pre>
</details>

<style>
  summary {
      cursor: pointer;
  }
  code {
      background-color: var(--base);
      padding: 5px 10px;
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