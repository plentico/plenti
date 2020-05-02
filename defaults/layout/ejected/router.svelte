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
      // Check if there is a 404 data source.
      node = getNode("/404");
      if (node === undefined) {
        // If no 404.json data source exists, pass placeholder values.
        node = {
          "path": "/404",
          "type": "404",
          "filename": "404.json",
          "fields": {}
        }
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

  const router = Navaid('/', () => {
    import('../content/404.js')
      .then(draw)
      .catch(err => {
        console.log("Add a '/layout/content/404.svelte' file to handle Page Not Found errors.");
        console.log("If you want to pass data to your 404 component, you can also add a '/content/404.json' file.");
        console.log(err);                                                                                           
      });
  });

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
