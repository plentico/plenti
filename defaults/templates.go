package defaults

var Templates = map[string][]byte{
	"/templates/layouts/html.js": []byte(`import React from 'react';
import ReactDOM from 'react-dom';

const title = "Home | Plenti";
const heading = "Welcome to plenti!";
const desc = "Your HTML page has been hydrated and you've enabled React with Webpack and Babel!";

ReactDOM.hydrate(
  <html>
    <head>
      <meta charset="utf-8" />
      <meta name="viewport" content="width=device-width" />
      <title>{title}</title>
    </head>
    <body>
      <h1>{heading}</h1>
      <p>{desc}</p>
      <p><a href="/about">About us</a>.</p>
    </body>
  </html>,
	document.getElementsByName('html')[0].value
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
