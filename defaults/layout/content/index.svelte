<script>
	export let title, intro, components, allContent;
	import Grid from '../components/grid.svelte';
  	import Uses from "../components/template.svelte";
</script>

<h1>{title}</h1>

<section id="intro">
	{#each intro as paragraph}
		<p>{@html paragraph}</p>
	{/each}
</section>

<div>
	<h3>Recent blog posts:</h3>
	<Grid items={allContent} filter="blog" />
	<br />
</div>

<Uses type="index" />