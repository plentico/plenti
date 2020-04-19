import svelte from 'svelte/compiler.js';
import path from 'path';
import fs from 'fs';

// Get the arguments from Go command execution.
const args = process.argv.slice(2)

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

// Return values to write files in Go.
//console.log(js.code);
//console.log("!plenti-split!");
//console.log(css.code);