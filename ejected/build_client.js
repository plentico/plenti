import svelte from 'svelte/compiler.js';

// The "component" variable gets injected by client.go.

// Create component JS that can run in the browser.
export let { js, css } = svelte.compile(component, {
	css: false
});

// Return the JS and CSS object.
(() => { js, css })();