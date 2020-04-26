<script>
	export let title, intro, allNodes;
	import Grid from '../components/grid.svelte';
</script>

<h1>{title}</h1>

{#if intro}<section id="intro"><p>{@html intro.slogan}</p></section>{/if}

<div>
	<h3>Recent blog posts:</h3>
	<Grid items={allNodes} filter="blog" />
	<br />
</div>

<details>
  <summary>Uses the "Index" template</summary>
  <pre><code>layout/content/index.svelte</code></pre>
</details>
