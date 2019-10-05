package defaults

var Templates = map[string][]byte{
	"/templates/layouts/html.js": []byte(`import React from 'react';
import ReactDOM from 'react-dom';

const title = 'React with Webpack and Babel';
ReactDOM.render(
  <div>{title}</div>,
  document.getElementById('app')
);`),
	"/templates/layouts/head.js": []byte(`<Thing>
	Placeholder..
</Thing>`),
	"/templates/layouts/header.js": []byte(`<Thing>
	Placeholder..
</Thing>`),
	"/templates/layouts/footer.js": []byte(`<Thing>
	Placeholder..
</Thing>`),
	"/templates/routed/pages.js": []byte(`<Thing>
	Placeholder..
</Thing>`),
}
