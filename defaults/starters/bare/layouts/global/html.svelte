<script>
  import Head from './head.svelte';

  export let content, layout, env;
</script>

<html lang="en">
  <Head title={content.filename} {env} />
  <body>
    <svelte:component this={layout} {...content.fields} />
  </body>
</html>