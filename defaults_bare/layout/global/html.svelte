<script>
  import Head from './head.svelte';

  export let route, content, allContent;
</script>

<html lang="en">
<Head title={content.filename} />
<body>
  <svelte:component this={route} {...content.fields} {allContent} />
</body>
</html>