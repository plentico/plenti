import Router from './router.svelte';

let target = document.querySelector('#hydrate-plenti').parentNode;

new Router({
  target: target,
  hydrate: true
});