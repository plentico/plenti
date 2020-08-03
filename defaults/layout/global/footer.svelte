<script>
  export let allContent;
  import { makeTitle } from '../scripts/make_title.svelte';
</script>

<footer>
  <div class="container">
    <span>All content:</span>
    {#each allContent as content}
      <a href="{content.path}">{makeTitle(content.filename)}</a>
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
