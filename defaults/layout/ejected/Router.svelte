<Html {Route} {node} {allNodes} />

  <script>
    import Navaid from 'navaid';
    import Content from './Content.js';
    import { onDestroy } from 'svelte';
    import Html from '../global/Html.svelte';
  
    let Route, node, allNodes;
  
    let uri = location.pathname;
    node = Content.getNode(uri);
    allNodes = Content.getAllNodes();
  
    function draw(m) {
      Route = m.default;
      window.scrollTo(0, 0);
    }
  
    function track(obj) {
      uri = obj.state || obj.uri;
      if (window.ga) ga.send('pageview', { dp:uri });
  
      node = Content.getNode(uri);
      allNodes = Content.getAllNodes();
    }
  
    addEventListener('replacestate', track);
    addEventListener('pushstate', track);
    addEventListener('popstate', track);
  
    const router = Navaid('/')
      .on('/', () => import('../content/Index.svelte').then(draw))
      .on('/:slug', () => import('../content/Pages.svelte').then(draw))
      .listen();
  
    onDestroy(router.unlisten);
  </script>