import Router from './router.svelte';
import allContent from './content.js';
import * as allLayouts from './layouts.js';
import { env } from './env.js';

let path = location.pathname;
let params = new URLSearchParams(location.search);
let layout, content;

const contentLookup = (path, trailingSlash = "") => {
  return allContent.find(content => content.path + trailingSlash == path); 
}

const makeRelativePath = path => { 
  // If first character is a forward slash and we're not on the homepage,
  // remove it before doing the content lookup. Do this recursively in case
  // multiple forward slashes are at the beginning of the path.
  return path.charAt(0) === "/" && path !== "/" ? makeRelativePath(path.substring(1)) : path;
}

const makeRootRelativePath = path => { 
  // Add a leading forward slash.
  return "/" + path;
}

export const getContent = path => {
  // Convert dot shorthand to slash when used for homepage links using base element.
  path = path === "." ? "/" : path;
  // Remove baseurl from beginning of path if it exists.
  path = path.replace(new RegExp('^\/?' + env.baseurl, 'i'),"");
  // Lookup content path with and without leading and trailing slashes.
  return contentLookup(path) ??
         contentLookup(makeRelativePath(path)) ??
         contentLookup(makeRootRelativePath(path)) ??
         contentLookup(path, "/") ??
         contentLookup(makeRelativePath(path), "/") ??
         contentLookup(makeRootRelativePath(path), "/")
}

content = getContent(path);

import('../content/' + content.type + '.js').then(r => {
  layout = r.default;
  new Router({
    target: document,
    hydrate: true,
    props: {
      path: path,
      params: params,
      layout: layout,
      content: content,
      allContent: allContent,
      allLayouts: allLayouts,
      env: env
    }
  });
}).catch(e => console.log(e));