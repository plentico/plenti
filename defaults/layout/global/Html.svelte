<script>
  import Head from './Head.svelte';
import Nav from './Nav.svelte';

export let Route, node, allNodes;

  const makeTitle = filename => {
  if (filename == '_index.json') {
    return 'Home';
  } else if (filename) {
    // Remove file extension.
    filename = filename.split('.').slice(0, -1).join('.');
    // Convert underscores and hyphens to spaces.
    filename = filename.replace(/_|-/g,' ');
    // Capitalize first letter of each word.
    filename = filename.split(' ').map((s) => s.charAt(0).toUpperCase() + s.substring(1)).join(' ');
  } 
  return filename;
  }
</script>

<html lang="en">
<Head title={makeTitle(node.filename)} />
<body>
  <Nav />
  <main>
    <svelte:component this={Route} {...node.fields} {allNodes} />
  </main>
</body>
</html>