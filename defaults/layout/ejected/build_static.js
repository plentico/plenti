import 'svelte/register.js';
import relative from 'require-relative';

// Get the arguments from command execution.
let args = process.argv.slice(2)

// args[0] is the path to /layout/global/html.svelte.
const component = relative(args[0], process.cwd()).default;

// args[1] is the path to /layout/content .svelte files.
const route = relative(args[1], process.cwd()).default;

// args[2] is the props being passed.
args[2].Route = route; // Add the correct component class instance.

// Create the static HTML and CSS.
export let { html, css } = component.render(args[1]);
