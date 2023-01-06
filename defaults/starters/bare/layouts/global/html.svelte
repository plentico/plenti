<script>
  import Head from './head.svelte';

  export let content, layout;
</script>

<html lang="en">
<Head title={content.filename} />
<body>
  <svelte:component this={layout} {...content.fields} />
</body>
</html>