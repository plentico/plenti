<Html {route} {content} {allContent} {allComponents} />

<script>
  import Navaid from 'navaid';
  import contentSource from './content.js';
  import Html from '../global/html.svelte';

  export let uri, route, content, allContent, allComponents;

  const getContent = (uri, trailingSlash = "") => {
    return contentSource.find(content => content.path + trailingSlash == uri);
  }

  function draw(m) {
    content = getContent(uri);
    if (content === undefined) {
      // Check if there is a 404 data source.
      content = getContent("/404");
      if (content === undefined) {
        // If no 404.json data source exists, pass placeholder values.
        content = {
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

  const handle404 = () => {
    import('../content/404.js')
      .then(draw)
      .catch(err => {
        console.log("Add a '/layout/content/404.svelte' file to handle Page Not Found errors.");
        console.log("If you want to pass data to your 404 component, you can also add a '/content/404.json' file.");
        console.log(err);
      });
  }

  const router = Navaid('/', handle404);

  allContent.forEach(content => {
    router.on(content.path, () => {
      import('../content/' + content.type + '.js').then(draw).catch(handle404);
    });

  });

  router.listen();

  // Check if the url visited ends in a trailing slash (besides the homepage).
  if (uri.length > 1 && uri.slice(-1) == "/") {
    // Redirect to the same path without the trailing slash.
    router.route(content.path, false);
  }

</script>
