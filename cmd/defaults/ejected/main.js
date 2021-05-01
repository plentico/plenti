import Router from './router.svelte';
import allContent from './content.js';
import * as allLayouts from './layouts.js';
import { local, baseurl } from './variables.js';

let uri = location.pathname;
let layout, content;

const getContent = (uri, trailingSlash = "") => {
  return allContent.find(content => content.path + trailingSlash == uri); 
}

const makeRelativeUri = uri => { 
  // If first character is a forward slash and we're not on the homepage,
  // remove it before doing the content lookup. Do this recursively in case
  // multiple forward slashes are at the beginning of the path.
  return uri.charAt(0) === "/" && uri !== "/" ? makeRelativeUri(uri.substring(1)) : uri;
}

const makeRootRelativeUri = uri => { 
  return "/" + uri;
}

const uriCombos = uri => {
  return getContent(uri) ??
         getContent(makeRelativeUri(uri)) ??
         getContent(makeRootRelativeUri(uri)) ??
         getContent(uri, "/") ??
         getContent(makeRelativeUri(uri), "/") ??
         getContent(makeRootRelativeUri(uri), "/")
}

//content = getContent(uri) != undefined ? getContent(uri) : getContent(uri, "/");
console.log(uri);
//content = getContent(uri) ?? getContent(makeRelativeUri(uri)) ?? getContent(uri, "/");
content = uriCombos(uri);
console.log(content);

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