<Html {route} {node} {allNodes} />

<script>
  import Navaid from 'navaid';
  import nodes from './nodes.js';
  import Html from '../global/html.svelte';

  let route, node, allNodes;

  const getNode = (uri, trailingSlash = "") => {
    return nodes.find(node => node.path + trailingSlash == uri);
  }

  let uri = location.pathname;
  node = getNode(uri);
  if (node === undefined) {
    node = getNode(uri, "/");
  }
  allNodes = nodes;

  function draw(m) {
    node = getNode(uri);
    if (node === undefined) {
      node = {
        "path": "/404",
        "type": "404",
        "filename": "404.json",
        "fields": {}
      }
    }
    route = m.default;
    window.scrollTo(0, 0);
  }

  function track(obj) {
    uri = obj.state || obj.uri;
  }

  addEventListener('replacestate', track);
  addEventListener('pushstate', track);
  addEventListener('popstate', track);

  const router = Navaid('/', () => import('../content/404.js').then(draw));

  allNodes.forEach(node => {
    router.on(node.path, () => {
      // Check if the url visited ends in a trailing slash (besides the homepage).
      if (uri.length > 1 && uri.slice(-1) == "/") {
        // Redirect to the same path without the trailing slash.
        router.route(node.path, false);
      } else {
        import('../content/' + node.type + '.js').then(draw);
      }
    });

  });

  router.listen();

</script>
