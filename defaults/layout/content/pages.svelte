<script>
	export let title, description, author;
</script>

<h1>{title}</h1>
<p><em>Page template</em></p>
{#if description}
  <div>
    {#each description as paragraph}
      <p>{@html paragraph}</p>
    {/each}
  </div>
{/if}
{#if author}
  <p>- <em>{author}</em></p>
{/if}

<p><a href="/">Back home</a></p>
