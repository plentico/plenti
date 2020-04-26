<script>
	export let title, body, author, date;
</script>

<h1>{title}</h1>

<p><em>{#if author}Written by {author}{/if}{#if date}&nbsp;on {date}{/if}</em></p>

{#if body}
  <div>
    {#each body as paragraph}
      <p>{@html paragraph}</p>
    {/each}
  </div>
{/if}

<details>
  <summary>Uses the "Blog" template</summary>
  <pre><code>layout/content/blog.svelte</code></pre>
</details>

<p><a href="/">Back home</a></p>
