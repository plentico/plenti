import svelte from 'svelte/compiler.js';

// Get the arguments from command execution.
const args = process.argv.slice(2)

// Create component JS that can run in the browser.
export let { js, css } = svelte.compile(args[0], {
	css: false
});
