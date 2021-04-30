<script>
  export let title, description, content;
  import Uses from "../components/source.svelte";
</script>

<h1>{title}</h1>

<div>
  {#each description as paragraph}
    <p>{@html paragraph}</p>
  {/each}
</div>

<Uses {content} />

<p><a href=".">Back home</a></p>