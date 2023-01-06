<script>
  export let currentPage, totalPages;
</script>

<ul>
  {#if currentPage > 1}
    <li><a href=".">first</a></li>
    <li><a href="{currentPage - 1}">prev</a></li>
  {:else}
    <li><span>first</span></li>
    <li><span>prev</span></li>
  {/if}
  {#each [3,2,1] as i}
    {#if currentPage - i > 0}
      <li><a href="{currentPage - i}">{currentPage - i}</a></li>
    {/if}
  {/each}
  <li><span>{currentPage}</span></li>
  {#each Array(3) as _, i}
    {#if currentPage + (i+1) <= totalPages}
      <li><a href="{currentPage + (i+1)}">{currentPage + (i+1)}</a></li>
    {/if}
  {/each}
  {#if currentPage < totalPages}
    <li><a href="{currentPage + 1}">next</a></li>
    <li><a href="{totalPages}">last</a></li>
  {:else}
    <li><span>next</span></li>
    <li><span>last</span></li>
  {/if}
</ul>

<style>
  ul {
    display: flex;
    max-width: 600px;
    margin: 20px auto;
  }
  li {
    flex-grow: 1;
    list-style: none;
  }
  a {
    text-align: center;
    padding: 5px 20px;
    border-radius: 5px;
    background: #f0efef;
  }
  span {
    text-align: center;
    padding: 5px 20px;
    color: gray;
  }
</style>