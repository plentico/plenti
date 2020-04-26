<script>
	export let title, description;
</script>

<h1>{title}</h1>

{#if description}
  <div>
    {#each description as paragraph}
      <p>{@html paragraph}</p>
    {/each}
  </div>
{/if}

<details>
  <summary>Uses the "Pages" template</summary>
  <pre><code>layout/content/pages.svelte</code></pre>
</details>

<p><a href="/">Back home</a></p>
