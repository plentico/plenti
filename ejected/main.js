import Router from './router.svelte';

const replaceContainer = Component => {
  const frag = document.createDocumentFragment();
  const dom = document.querySelector('#hydrate-plenti');
  const component = new Component( Object.assign( {}, { target: dom }, { target: frag } ));
  dom.replaceWith(frag);
}

replaceContainer(Router);