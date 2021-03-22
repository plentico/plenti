<script>
  import Head from './head.svelte';

  export let content, layout, allContent;
</script>

<html lang="en">
<Head title={content.filename} />
<body>
  <svelte:component this={layout} {...content.fields} {allContent} />
</body>
</html>