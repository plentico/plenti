<script>
  export let allNodes;
  import { makeTitle } from '../scripts/make_title.svelte';
</script>

<footer>
  <div class="container">
    <span>All nodes:</span>
    {#each allNodes as node}
      <a href="{node.path}">{makeTitle(node.filename)}</a>
    {/each}
  </div>
</footer>

<style>
  footer {
    min-height: 200px;
    display: flex;
    align-items: center;
    background-color: var(--base-dark);
    margin-top: 100px;
  }
  span {
    color: var(--primary);
    font-weight: bold;
  }
  a {
    color: white;
    text-decoration: none;
    margin-left: 10px;
  }
</style>
