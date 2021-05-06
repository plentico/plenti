import Router from './router.js';
import allContent from './content.js';
import * as allLayouts from './layouts.js';
import { env } from './env.js';

let uri = location.pathname;
let layout, content;

const contentLookup = (uri, trailingSlash = "") => {
  return allContent.find(content => content.path + trailingSlash == uri); 
}

const makeRelativeUri = uri => { 
  // If first character is a forward slash and we're not on the homepage,
  // remove it before doing the content lookup. Do this recursively in case
  // multiple forward slashes are at the beginning of the path.
  return uri.charAt(0) === "/" && uri !== "/" ? makeRelativeUri(uri.substring(1)) : uri;
}

const makeRootRelativeUri = uri => { 
  // Add a leading forward slash.
  return "/" + uri;
}

export const getContent = uri => {
  // Convert dot shorthand to slash when used for homepage links using base element.
  uri = uri === "." ? "/" : uri;
  // Lookup content path with and without leading and trailing slashes.
  return contentLookup(uri) ??
         contentLookup(makeRelativeUri(uri)) ??
         contentLookup(makeRootRelativeUri(uri)) ??
         contentLookup(uri, "/") ??
         contentLookup(makeRelativeUri(uri), "/") ??
         contentLookup(makeRootRelativeUri(uri), "/")
}

content = getContent(uri);

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
      env: env
    }
  });
}).catch(e => console.log(e));