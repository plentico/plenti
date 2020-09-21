import Router from './router.svelte';

let target = document.querySelector('html').parentNode;

new Router({
  target: target,
  hydrate: true
});