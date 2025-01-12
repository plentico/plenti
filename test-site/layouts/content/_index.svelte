<script>
	export let title, intro, blog, source, content, allContent;
	import Grid from '../components/grid.svelte';
	import Uses from "../components/source.svelte";
	import Pager from "../components/pager.svelte";

	$: currentPage = content.pager ? content.pager : 1;
	let postsPerPage = 3;
	let allPosts = allContent.filter(content => content.type == "blog");
	let totalPosts = allPosts.length;
	let totalPages = Math.ceil(totalPosts / postsPerPage);
	$: postRangeHigh = currentPage * postsPerPage;
	$: postRangeLow = postRangeHigh - postsPerPage;
</script>

<h1>{title}</h1>

<section id="intro">
	{#each intro as paragraph}
		<p>{@html paragraph}</p>
	{/each}
</section>

{#if blog}
	<div>
		<h3>Recent blog posts:</h3>
		<Grid items={allPosts} {postRangeLow} {postRangeHigh} />
		<br />
	</div>
	<Pager {currentPage} {totalPages} />	
{/if}

{#if source}
	<Uses {content} {source} />	
{/if}