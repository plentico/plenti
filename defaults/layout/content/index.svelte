<script>
	export let title, intro, components, allNodes;
	import Grid from '../components/grid.svelte';
	import { loadComponent } from '../scripts/load_component.svelte';
</script>

<h1>{title}</h1>

<section id="intro">
	<p>{@html intro.slogan}</p>
</section>

<div>
	<h3>Recent blog posts:</h3>
	<Grid items={allNodes} filter="blog" />
	<br />
</div>

{#if components}
	{#each components as { component, fields }}
		{#await loadComponent(component)}
		{:then compClass}
			<svelte:component this="{compClass}" {...fields} />
		{:catch error}
		{/await}
	{/each}
{/if}