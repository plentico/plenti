import svelte from 'svelte/compiler.js';
import 'svelte/register.js';
import Module from 'module';
import path from 'path';
import fs from 'fs';

// Get the arguments from Go command execution.
const args = process.argv.slice(2)

// -----------------
// Helper Functions:
// -----------------

// Create any missing sub folders.
const ensureDirExists = filePath => {
	let dirname = path.dirname(filePath);
	if (fs.existsSync(dirname)) {
		return true;
	}
	ensureDirExists(dirname);
	fs.mkdirSync(dirname);
}

// Concatenates HTML strings together.
const injectString = (order, content, element, html) => {
	if (order == 'prepend') {
		return html.replace(element, content + element);
	} else if (order == 'append') {
		return html.replace(element, element + content);
	}
};

// -----------------------
// Start client SPA build:
// -----------------------

let clientBuildStr = JSON.parse(args[0]);

clientBuildStr.forEach(arg => {

	let layoutPath = path.join(path.resolve(), arg.layoutPath)
	let component = fs.readFileSync(layoutPath, 'utf8');

	// Create component JS that can run in the browser.
	let { js, css } = svelte.compile(component, {
		css: false,
		hydratable: true
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

let staticBuildStr = JSON.parse(args[1]);
let allContent = JSON.parse(args[2]);

// Create the component that wraps all content.
let htmlWrapper = path.join(path.resolve(), 'layout/global/html.svelte')
let root = new Module();
let component = root.require(htmlWrapper).default;

staticBuildStr.forEach(arg => {

	let componentPath = path.join(path.resolve(), arg.componentPath);
	let destPath = path.join(path.resolve(), arg.destPath);

	// Set route used in svelte:component as "this" value.
	const route = root.require(componentPath).default;

	// Set props so component can access field values, etc.
	let props = {
		route: route,
		content: arg.content,
		allContent: allContent
	};

	// Create the static HTML and CSS.
	let { html, css } = component.render(props);

	// Inject Style.
	let style = "<style>" + css.code + "</style>";
	html = injectString('prepend', style, '</head>', html);
	// Inject SPA entry point.
	let entryPoint = '<script type="module" src="https://unpkg.com/dimport?module" data-main="/spa/ejected/main.js"></script><script nomodule src="https://unpkg.com/dimport/nomodule" data-main="/spa/ejected/main.js"></script>';
	html = injectString('prepend', entryPoint, '</head>', html);

	// Write .html file to filesystem.
  	ensureDirExists(destPath);
	fs.promises.writeFile(destPath, html);
	  
});