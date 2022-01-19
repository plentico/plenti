<script>
  export let title, description, content, cms, login;
  import Uses from "../components/source.svelte";
</script>

<h1>{title}</h1>

<div>
  {#each description as paragraph}
    <p>{@html paragraph}</p>
  {/each}
</div>

{#if cms}
  <p>Try it out: <button on:click={login}>Login</button></p>
{/if}

<Uses {content} />

<p><a href=".">Back home</a></p>