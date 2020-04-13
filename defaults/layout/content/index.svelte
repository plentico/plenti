<script>
	export let name;
	export let allNodes;
  import Grid from '../components/grid.svelte';
</script>

<h1>{name}</h1>
<section id="intro">
	<p>Visit the <a href="https://svelte.dev/tutorial">Svelte tutorial</a> to learn how to build Svelte apps.</p>
	<h3>Recent blog posts:</h3>
  <Grid items={allNodes} filter="blog" />
</section>
