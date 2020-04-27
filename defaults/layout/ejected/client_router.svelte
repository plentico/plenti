<Html {Route} {node} {allNodes} />

<script>
  import Navaid from 'navaid';
  import DataSource from './data_source.js';
  import Html from '../global/html.svelte';

  let Route, node, allNodes;

  let uri = location.pathname;
  node = DataSource.getNode(uri);
  allNodes = DataSource.getAllNodes();

  function draw(m) {
    Route = m.default;
    window.scrollTo(0, 0);
  }

  function track(obj) {
    uri = obj.state || obj.uri;
    if (window.ga) ga.send('pageview', { dp:uri });

    node = DataSource.getNode(uri);
    allNodes = DataSource.getAllNodes();
  }

  addEventListener('replacestate', track);
  addEventListener('pushstate', track);
  addEventListener('popstate', track);

  const router = Navaid('/', () => import('../global/404.svelte').then(draw));

  allNodes.forEach(node => {
    router.on(node.path, () => import('../content/' + node.type + '.svelte').then(draw)).listen();
  });

</script>
