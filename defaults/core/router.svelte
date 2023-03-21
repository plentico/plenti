<Html
  {path}
  {params}
  {content}
  {layout}
  {allContent}
  {allLayouts}
  {env}
  {user}
  {shadowContent}
/>

<script>
  import Html from '../layouts/global/html.svelte';
  import Navaid from 'navaid';
  import allContent from '../generated/content.js';
  import * as allLayouts from '../generated/layouts.js';
  import { env } from '../generated/env.js';
  import { user } from './cms/auth.js';
  import allDefaults from '../generated/defaults.js';

  let path = location.pathname;
  let params = new URLSearchParams(location.search);
  // Load data-content-filepath attribute from HTML
  let content = allContent.find(c => c.filepath === document.documentElement.dataset.contentFilepath);
  let layout;
  let shadowContent = {};

  if ($user.isBeingAuthenticated) { 
    $user.finishAuthentication(params);
  }

  function track(obj) {
    path = obj.state || obj.uri || location.pathname;
    params = new URLSearchParams(location.search);
  }

  addEventListener('replacestate', track);
  addEventListener('pushstate', track);
  addEventListener('popstate', track);

  const handle404 = () => {
    // Check if there is a 404 data source.
    content = allContent.find(c => c.filepath === "content/404.json");
    if (content === undefined) {
      // If no 404.json data source exists, pass placeholder values.
      content = {
        "path": "/404",
        "type": "404",
        "filename": "404.json",
        "fields": {}
      }
    }
    import('../layouts/content/404.js')
      .then(component => {
        layout = component.default;
      })
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
    let baseurl = env.baseurl ? env.baseurl : '/';
    if (location.pathname !== baseurl) {
      return false;
    }

    if (location.hash.startsWith('#add/') && $user.isAuthenticated) {
      const [type, filename] = location.hash.substring('#add/'.length).split('/');
      const defaultContent = allDefaults.find(defaultContent => defaultContent.type == type);

      if (type && filename && defaultContent) {
        import('../layouts/content/' + type + '.js').then(m => {
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
  allContent.forEach(currentContent => {
    router.on(env.baseurl + currentContent.path, () => {
      // Override with hash location if one is found.
      if (navigateHashLocation()) {
        return;
      }

      import('../layouts/content/' + currentContent.type + '.js')
        .then(component => {
          if (content.filepath !== currentContent.filepath) {
            window.scrollTo(0, 0);
          }
          content = currentContent;
          layout = component.default;
        })
        .catch(handle404);
    });
  });
  router.listen();
</script>
