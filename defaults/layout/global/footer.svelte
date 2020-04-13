<script>
  export let allNodes;
</script>
<footer>
  <div class="container">
    <span>All nodes:</span>
    {#each allNodes as node}
      {#if node.fields.title}
        <a href="{node.path}">{node.fields.title}</a>&nbsp;
      {/if}
    {/each}
  </div>
</footer>

<style>
  footer {
    min-height: 200px;
    display: flex;
    align-items: center;
    background-color: var(--base);
    margin-top: 100px;
  }
</style>
