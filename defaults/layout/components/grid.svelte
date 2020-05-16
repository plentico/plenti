<script>
  import { sortByDate } from '../scripts/sort_by_date.svelte';
  export let items, filter;
</script>

<div class="grid">
  {#each sortByDate(items) as item}
		{#if item.type == filter}
      <a class="grid-item" href="{item.path}">{item.fields.title}</a>
		{/if}
  {/each}
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    grid-column-gap: 10px;
    grid-row-gap: 10px;
  }
  .grid-item {
    display: flex;
    flex-grow: 1;
    height: 200px;
    align-items: center;
    justify-content: center;
    background: var(--primary);
    font-weight: bold;
    border-radius: 5px;
    color: white;
  }
  a::before {
    content: none;
  }
</style>
