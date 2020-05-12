<script>
  export let title, description;
  import Uses from "../components/template.svelte";
</script>

<h1>{title}</h1>

<div>
  {#each description as paragraph}
    <p>{@html paragraph}</p>
  {/each}
</div>

<Uses type="pages" />

<p><a href="/">Back home</a></p>