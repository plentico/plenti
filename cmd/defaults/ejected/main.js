import Router from './router.svelte';
import contentSource from './content.js';
import * as allComponents from './layout.js';

let uri = location.pathname;
let route, content, allContent;

const getContent = (uri, trailingSlash = "") => {
  return contentSource.find(content => content.path + trailingSlash == uri);
}

content = getContent(uri) != undefined ? getContent(uri) : getContent(uri, "/");
allContent = contentSource;

import('../content/' + content.type + '.js').then(r => {
  route = r.default;
  new Router({
    target: document,
    hydrate: true,
    props: {
      uri: uri,
      route: route,
      content: content,
      allContent: allContent,
      allComponents: allComponents
    }
  });
}).catch(e => console.log(e));