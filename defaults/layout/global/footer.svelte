<script>
  export let allNodes;
  import { makeTitle } from '../scripts/make_title.svelte';
</script>
<footer>
  <div class="container">
    <span>All nodes:</span>
    {#each allNodes as node}
      <a href="{node.path}">{makeTitle(node.filename)}</a>&nbsp;
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
