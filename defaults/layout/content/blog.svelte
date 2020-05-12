<script>
	export let title, body, author, date;
  import Uses from "../components/template.svelte";
</script>

<h1>{title}</h1>

<p><em>{#if author}Written by {author}{/if}{#if date}&nbsp;on {date}{/if}</em></p>

<div>
  {#each body as paragraph}
    <p>{@html paragraph}</p>
  {/each}
</div>

<Uses type="blog" />

<p><a href="/">Back home</a></p>
