<script>
  export let title, description, source, content, cms, user;
  import Uses from "../components/source.svelte";
</script>

<h1>{title}</h1>

<div>
  {#each description as paragraph}
    <p>{@html paragraph}</p>
  {/each}
</div>

{#if cms}
  <p>Try it out: <button on:click={$user.login}>Login</button></p>
{/if}

{#if source}
  <Uses {content} {source} />
{/if}

<p><a href=".">Back home</a></p>