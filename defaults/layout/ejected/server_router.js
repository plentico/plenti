import path from 'path';
import fs from 'fs';
import nodes from './nodes.js';
import 'svelte/register.js';
import relative from 'require-relative';
import svelte from 'svelte/compiler.js';

const injectString = (order, content, element, html) => {
	if (order == 'prepend') {
		return html.replace(element, content + element);
	} else if (order == 'append') {
		return html.replace(element, element + content);
	}
};

const ensureDirExists = filePath => {
	var dirname = path.dirname(filePath);
	if (fs.existsSync(dirname)) {
		return true;
	}
	ensureDirExists(dirname);
	fs.mkdirSync(dirname);
}
// Start client
let sPaths = []; 
sPaths.push('ejected/client_router.svelte');
sPaths.push('ejected/main.js');
sPaths.push('ejected/nodes.js');
sPaths.push('ejected/data_source.js');
sPaths.push('global/html.svelte');
sPaths.push('global/nav.svelte');
sPaths.push('global/head.svelte');
sPaths.push('global/footer.svelte');
sPaths.push('content/pages.svelte');
sPaths.push('content/index.svelte');
sPaths.push('content/blog.svelte');
sPaths.push('components/grid.svelte');
sPaths.push('scripts/make_title.svelte');
sPaths.forEach(sPath => {
  let extension = sPath.substring(sPath.lastIndexOf('.')+1, sPath.length);
  if (extension == 'js') {
    let sDest = 'public/spa/' + sPath;
		ensureDirExists(sDest);
    fs.copyFile('layout/' + sPath, sDest, (err) => {
        if (err) throw err;
        //console.log('File was copied to destination');
    });
    return;
  }
  let spaSourcePath = path.join(path.resolve(), 'layout/' + sPath);
	let spaSourceComponent = fs.readFileSync(spaSourcePath, 'utf8');
	let { js, css } = svelte.compile(spaSourceComponent, {
    css: false
	});
	let spaDestPath = 'public/spa/' + sPath.substr(0, sPath.lastIndexOf(".")) + ".js";
  js.code = js.code.replace(/\.svelte/g, '.js');
  js.code = js.code.replace(/from "svelte\/internal"\;/g, 'from "../web_modules/svelte/internal/index.js";');
  js.code = js.code.replace(/from "svelte"\;/g, 'from "../web_modules/svelte.js";');
  js.code = js.code.replace(/from "navaid"\;/g, 'from "../web_modules/navaid.js";');
	ensureDirExists(spaDestPath);
  if (css.code && css.code != 'null') {
    fs.appendFileSync('public/spa/bundle.css', css.code);
  }
	fs.promises.writeFile(spaDestPath, js.code);
});
// End client

nodes.forEach(node => {
  let sourcePath = path.join(path.resolve(), 'layout/content/' + node.type + '.svelte');
  let sourceComponent = fs.readFileSync(sourcePath, 'utf8');
  let index = node.filename == 'index.json' ? 'index' : '';
  let destPath = path.join(path.resolve(), 'public/' + node.path + index + ".html");
  let topLevelComponent = path.join(path.resolve(), 'layout/global/html.svelte');
  const route = relative(sourcePath, process.cwd()).default;
  let props = {
    Route: route,
    node: node,
    allNodes: nodes
  };
  // Create HTML file.
  const component = relative(topLevelComponent, process.cwd()).default;
  let { html, css } = component.render(props);
  // Inject Style.
  let style = "<style>" + css.code + "</style>";
  html = injectString('prepend', style, '</head>', html);
  // Inject SPA entry point.
  let entryPoint = '<script type="module" src="https://unpkg.com/dimport?module" data-main="/spa/ejected/main.js"></script><script nomodule src="https://unpkg.com/dimport/nomodule" data-main="/spa/ejected/main.js"></script>';
  html = injectString('prepend', entryPoint, '</head>', html);
  // Inject ID used to hydrate SPA.
  let hydrator = ' id="hydrate-plenti"';
  html = injectString('append', hydrator, '<html', html);
  // Write HTML files to filesystem.
  ensureDirExists(destPath);
  fs.promises.writeFile(destPath, html);
});
