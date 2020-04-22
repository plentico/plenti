import svelte from 'svelte/compiler.js';
import 'svelte/register.js';
import relative from 'require-relative';
import path from 'path';
import fs from 'fs';

// Get the arguments from Go command execution.
const args = process.argv.slice(2)

// -----------------------
// Start client SPA build:
// -----------------------

// Create any missing sub folders.
const ensureDirExists = filePath => {
	let dirname = path.dirname(filePath);
	if (fs.existsSync(dirname)) {
		return true;
	}
	ensureDirExists(dirname);
	fs.mkdirSync(dirname);
}

let buildStr = JSON.parse(args[0]);

buildStr.forEach(arg => {
	// Create component JS that can run in the browser.
	let { js, css } = svelte.compile(arg.component, {
		css: false
	});

	// Write JS to build directory.
	ensureDirExists(arg.destPath);
	fs.promises.writeFile(arg.destPath, js.code);

	// Write CSS to build directory.
	ensureDirExists(arg.stylePath);
	if (css.code && css.code != 'null') {
		fs.appendFileSync(arg.stylePath, css.code);
	}
});

// ------------------------
// Start static HTML build:
// ------------------------

/*
let wrapper = path.join(path.resolve(), 'layout/global/html.svelte')
const component = relative(wrapper, process.cwd()).default;

// args[1] is the path to /layout/content .svelte files.
const route = relative(args[1], process.cwd()).default;

// args[2] is the props being passed.
args[2].Route = route; // Add the correct component class instance.

// Create the static HTML and CSS.
let { html, css } = component.render(args[1]);
*/