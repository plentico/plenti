<script>
	export let allNodes;
	export let title, description;
</script>

<h1>{title}</h1>
<p>Pages template</p>
<a href="/">Back home</a>
<div>
<strong>Title:</strong><span>{title}</span>
<strong>Desc:</strong><span>{description}</span>
</div>

<h3>All nodes test:</h3>
{#each allNodes as node}
	<a href="{node.path}">{node.fields.title}</a>
{/each}