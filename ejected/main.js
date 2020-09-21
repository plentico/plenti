import Router from './router.svelte';

new Router({
  target: document.querySelector('#hydrate-plenti'),
  hydrate: true
});