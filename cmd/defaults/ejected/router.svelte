<Html {path} {params} {content} {layout} {allContent} {allLayouts} {env} {user} {adminMenu} {shadowContent} />

<script>
  import Navaid from 'navaid';
  import Html from '../global/html.svelte';
  import { getContent } from './main.js';

  export let path, params, content, layout, allContent, allLayouts, env;

  let shadowContent = {};

  function draw(m) {
    content = getContent(path); 
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
    layout = m.default;
    window.scrollTo(0, 0);
  }

  function track(obj) {
    path = obj.state || obj.uri || location.pathname;
    params = new URLSearchParams(location.search);
  }

  addEventListener('replacestate', track);
  addEventListener('pushstate', track);
  addEventListener('popstate', track);

  const handle404 = () => {
    import('../content/404.js')
      .then(draw)
      .catch(err => {
        console.log("Add a '/layouts/content/404.svelte' file to handle Page Not Found errors.");
        console.log("If you want to pass data to your 404 component, you can also add a '/content/404.json' file.");
        console.log(err);
      });
  }

  /**
   * @return {boolean} true if hash location found and navigated, false otherwise.
   */
  const navigateHashLocation = () => {
    if (location.pathname != '/') {
      return false;
    }

    if (location.hash.startsWith('#add/') && $user.isAuthenticated) {
      const [type, filename] = location.hash.substring('#add/'.length).split('/');
      const defaultContent = allDefaults.find(defaultContent => defaultContent.type == type);

      if (type && filename && defaultContent) {
        import('../content/' + type + '.js').then(m => {
          content = structuredClone(defaultContent);
          content.isNew = true;
          content.filename = filename + '.json';
          content.filepath = content.filepath.replace('_defaults.json', filename + '.json');
          layout = m.default;
        }).catch(handle404);
        return true;
      } else {
        // Page type not found or filename not specified.
        handle404();
        return true;
      }
    }

    return false;
  };

  const router = Navaid('/', handle404);
  allContent.forEach(content => {
    router.on((env.local ? '' : env.baseurl) + content.path, () => {
      // Override with hash location if one is found.
      if (navigateHashLocation()) {
        return;
      }

      import('../content/' + content.type + '.js').then(draw).catch(handle404);
    });
  });
  router.listen();

  // Git-CMS
  import adminMenu from './cms/admin_menu.svelte';
  import { user } from './cms/auth.js';
  import allDefaults from './defaults.js';
  if ($user.isBeingAuthenticated) { 
    $user.finishAuthentication(params);
  }
</script>
