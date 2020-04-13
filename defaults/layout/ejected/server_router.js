import path from 'path';
import fs from 'fs';
import nodes from './nodes.js';
import 'svelte/register.js';
import relative from 'require-relative';

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
  let style = `<style>${css.code}</style>`;
  html = injectString('prepend', style, '</head>', html);
  // Inject SPA entry point.
  let entryPoint = `
  <script type="module" src="https://unpkg.com/dimport?module" data-main="/.spa/main.js"></script>
  <script nomodule src="https://unpkg.com/dimport/nomodule" data-main="/.spa/main.js"></script>
	`;
  html = injectString('prepend', entryPoint, '</head>', html);
  // Inject ID used to hydrate SPA.
  let hydrator = ' id="hydrate-plenti"';
  html = injectString('append', hydrator, '<html', html);
  // Write HTML files to filesystem.
  ensureDirExists(destPath);
  fs.promises.writeFile(destPath, html);
});
