import Router from './client_router.svelte';

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/jim-service-worker.js')
  .then((reg) => {
    console.log('Service Worker registration succeeded.');
  }).catch((error) => {
    console.log('Service Worker registration failed with ' + error);
  });
} else {
  console.log('Service Workers not supported by browser')
}

const replaceContainer = function ( Component, options ) {
  const frag = document.createDocumentFragment();
  const component = new Component( Object.assign( {}, options, { target: frag } ));
  if (options.target) {
    options.target.replaceWith( frag );
  }
  return component;
}

const app = replaceContainer( Router, {
  target: document.querySelector( '#hydrate-plenti' ),
  props: {}
});

export default app;
