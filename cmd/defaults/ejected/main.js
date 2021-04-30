import Router from './router.svelte';
import allContent from './content.js';
import * as allLayouts from './layouts.js';
import { local, baseurl } from './variables.js';

let uri = location.pathname;
let layout, content;

const getContent = (uri, trailingSlash = "") => {
  uri = uri === "/" ? uri : uri.substring(1);
  console.log("main.js: " + uri);
  return allContent.find(content => content.path + trailingSlash == uri);
}

content = getContent(uri) != undefined ? getContent(uri) : getContent(uri, "/");

import('../content/' + content.type + '.js').then(r => {
  layout = r.default;
  new Router({
    target: document,
    hydrate: true,
    props: {
      uri: uri,
      layout: layout,
      content: content,
      allContent: allContent,
      allLayouts: allLayouts,
      local: local,
      baseurl: baseurl
    }
  });
}).catch(e => console.log(e));