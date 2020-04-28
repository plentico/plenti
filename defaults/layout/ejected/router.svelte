<Html {route} {node} {allNodes} />

<script>
  import Navaid from 'navaid';
  import nodes from './nodes.js';
  import Html from '../global/html.svelte';

  let route, node, allNodes;

  const getNode = uri => {
    return nodes.find(node => node.path == uri);
  }

  let uri = location.pathname;
  node = getNode(uri);
  allNodes = nodes;

  function draw(m) {
    route = m.default;
    window.scrollTo(0, 0);
  }

  function track(obj) {
    uri = obj.state || obj.uri;
    if (window.ga) ga.send('pageview', { dp:uri });

    node = getNode(uri);
    allNodes = nodes;
  }

  addEventListener('replacestate', track);
  addEventListener('pushstate', track);
  addEventListener('popstate', track);

  const router = Navaid('/', () => import('../global/404.js').then(draw));

  allNodes.forEach(node => {
    router.on(node.path, () => import('../content/' + node.type + '.js').then(draw)).listen();
  });

</script>
