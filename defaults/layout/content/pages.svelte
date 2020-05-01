<script>
	export let title, description;
</script>

<h1>{title}</h1>

<div>
  {#each description as paragraph}
    <p>{@html paragraph}</p>
  {/each}
</div>

<details>
  <summary>Uses the "Pages" template</summary>
  <pre><code>layout/content/pages.svelte</code></pre>
</details>

<p><a href="/">Back home</a></p>
