import Router from './router.svelte';
import allContent from '../generated/content.js';
import * as allLayouts from '../generated/layouts.js';
import { env } from '../generated/env.js';

// Load data-content-filepath attribute from HTML
let content = allContent.find(c => c.filepath === document.documentElement.dataset.contentFilepath);
let path = location.pathname;
let params = new URLSearchParams(location.search);

import('../layouts/content/' + content.type + '.js').then(r => {
  let layout = r.default;
  new Router({
    target: document,
    hydrate: true,
    props: {
      content: content,
      layout: layout,
      allContent: allContent,
      allLayouts: allLayouts,
      path: path,
      params: params,
      env: env
    }
  });
}).catch(e => console.log(e));