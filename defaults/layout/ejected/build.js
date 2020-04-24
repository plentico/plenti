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

let clientBuildStr = JSON.parse(args[0]);

clientBuildStr.forEach(arg => {
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

//console.log(args[1]);
//console.log("\n\n");
//console.log(args[2]);

let staticBuildStr = JSON.parse(args[1]);
let allNodes = JSON.parse(args[2]);

let htmlWrapper = path.join(path.resolve(), 'layout/global/html.svelte')
const component = relative(htmlWrapper, process.cwd()).default;

staticBuildStr.forEach(arg => {

	let componentPath = path.join(path.resolve(), arg.componentPath);
	let destPath = path.join(path.resolve(), arg.destPath);

	const route = relative(componentPath, process.cwd()).default;

	let props = {
		Route: route,
		node: arg.node,
		allNodes: allNodes
	};

	// Create the static HTML and CSS.
	let { html, css } = component.render(props);

	// Write .html file to filesystem.
  	ensureDirExists(destPath);
	fs.promises.writeFile(destPath, html);
	  
});