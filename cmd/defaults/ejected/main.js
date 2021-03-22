import Router from './router.svelte';
import contentSource from './content.js';
import * as allLayouts from './layouts.js';

let uri = location.pathname;
let layout, content, allContent;

const getContent = (uri, trailingSlash = "") => {
  return contentSource.find(content => content.path + trailingSlash == uri);
}

content = getContent(uri) != undefined ? getContent(uri) : getContent(uri, "/");
allContent = contentSource;

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
      allLayouts: allLayouts
    }
  });
}).catch(e => console.log(e));