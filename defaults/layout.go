package defaults

// Layout : default site structure
var Layout = map[string][]byte{
	"/layout/ejected/main.js": []byte(`import Router from './Router.svelte';

  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/jim-service-worker.js')
    .then((reg) => {
      console.log('Service Worker registration succeeded.');
    }).catch((error) => {
      console.log('Service Worker registration failed with ' + error);
    });
  } else {
    console.log('Service Workers not supported by browser')
  }
  
  const replaceContainer = function ( Component, options ) {
    const frag = document.createDocumentFragment();
    const component = new Component( Object.assign( {}, options, { target: frag } ));
    if (options.target) {
      options.target.replaceWith( frag );
    }
    return component;
  }
  
  const app = replaceContainer( Router, {
    target: document.querySelector( '#hydrate-plenti' ),
    props: {}
  });
  
  export default app;`),
	"/layout/ejected/Router.svelte": []byte(`<Html {Route} {node} {allNodes} />

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
  </script>`),
	"/layout/ejected/Content.js": []byte(`import nodes from './nodes.js';

  class Content {
  
    constructor() {}
  
    static getNode(uri) {
      let content;
      nodes.map(node => {
        if (node.path == uri) {
          content = node;
        }
      });
      return content ? content : '';
    }
  
    static getAllNodes() {
      let content = nodes.map(node => {
        return node;
      });
      return content;
    }
  }
  
  export default Content;
  `),
	"/layout/ejected/nodes.js": []byte(`const content = [
    {
        "path": "/blog/post1",
        "type": "blog",
        "filename": "post1.json",
        "fields": {
            "title": "Post 1",
            "description": "First blog post."
        }
    },
    {
        "path": "/blog/post2",
        "type": "blog",
        "filename": "post2.json",
        "fields": {
            "title": "Post 2",
            "description": "Second blog post."
        }
    },
    {
        "path": "/blog/post3",
        "type": "blog",
        "filename": "post-3_has_a_long_filename.json",
        "fields": {
            "title": "Post 3",
            "description": "Third of the blog posts."
        }
    },
    {
        "path": "/about",
        "type": "pages",
        "filename": "about.json",
        "fields": {
            "title": "About Page",
            "description": "This is the about page"
        }
    },
    {
        "path": "/anything",
        "type": "pages",
        "filename": "anything.json",
        "fields": {
            "title": "Anything!",
            "description": "The amazing anything page..."
        }
    },
    {
        "path": "/",
        "type": "",
        "filename": "_index.json",
        "fields": {
            "name": "Plenti"
        }
    }
];

export default content;
  `),
	"/layout/global/Head.svelte": []byte(`<script>
  export let title;
</script>

<head>
  <meta charset='utf-8'>
  <meta name='viewport' content='width=device-width,initial-scale=1'>

  <title>{ title }</title>

  <link rel='icon' type='image/png' href='/favicon.png'>
  <link rel='stylesheet' href='/build/bundle.css'>
</head>`),
	"/layout/global/Html.svelte": []byte(`<script>
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
</html>`),
	"/layout/global/Nav.svelte": []byte(`<nav>
  <span id="brand"><a href="/">Home</a></span>
  <a href="/about">About</a>
  <a href="/anything">Anything</a>
</nav>`),
	"/layout/content/Index.svelte": []byte(`<script>
	export let name;
	export let allNodes;
</script>

<section id="intro">
	<img alt="plenti logo" src="/build/plenti.svg" />
	<h1>{name}</h1>
	<p>Visit the <a href="https://svelte.dev/tutorial">Svelte tutorial</a> to learn how to build Svelte apps.</p>
	<h3>Recent blog posts:</h3>
	{#each allNodes as node}
		{#if node.type == 'blog'}
			<a href="{node.path}">{node.fields.title}</a>
			<br />
		{/if}
	{/each}
</section>`),
	"/layout/content/Pages.svelte": []byte(`<script>
	export let allNodes;
	export let title, description;
</script>

<h1>{title}</h1>
<p>Pages template</p>
<a href="/">Back home</a>
<div>
<strong>Title:</strong><span>{title}</span>
<strong>Desc:</strong><span>{description}</span>
</div>

<h3>All nodes test:</h3>
{#each allNodes as node}
	<a href="{node.path}">{node.fields.title}</a>
{/each}
  `),
}
