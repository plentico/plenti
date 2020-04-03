<script>
	export let name;
	export let allNodes;
</script>

<section id="intro">
	<img alt="plenti logo" src="/build/plenti.svg" />
	<h1>{name}</h1>
	<p>Visit the <a href="https://svelte.dev/tutorial">Svelte tutorial</a> to learn how to build Svelte apps.</p>
	<h3>Recent blog posts:</h3>
	{#each allNodes as node}
		{#if node.type == 'blog'}
			<a href="{node.path}">{node.fields.title}</a>
			<br />
		{/if}
	{/each}
</section>