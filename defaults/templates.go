package defaults

var Templates = map[string][]byte{
	"/templates/entry.js": []byte(`import React from 'react';
import ReactDOM from 'react-dom';
import HTML from './layouts/html';

ReactDOM.hydrate(
  <HTML />,
  document.getElementsByTagName('html')[0]
);`),
	"/templates/layouts/html.js": []byte(`import React, { Component } from 'react';
import { BrowserRouter, Route, Link, Switch } from 'react-router-dom';
import Head from './head';
import Router from '../routes';

const title = "Home | Plenti";

class HTML extends Component {
  render() {
    return (
      <html lang="en" dir="ltr">
        <Head title={title} />
        <body>
          <Router />
        </body>
      </html>
    );
  }
}

export default HTML;`),
	"/templates/routes.js": []byte(`import React, { Component } from 'react';
import { BrowserRouter, Route, Link, Switch } from 'react-router-dom';

class Router extends Component {
  render() {
    return (
      <BrowserRouter>
        <Switch>
          <Route exact path="/" component={Home} />
          <Route path="/about" component={About} />
          <Route component={NoMatch} />
        </Switch>
      </BrowserRouter>
    );
  }
}

class Home extends Component {
  render() {
    return (
      <div>
        <h1>Home</h1>
        <p>This is the home page. <Link to="/about">About us</Link>.</p>
      </div>
    );
  }
}

class About extends Component {
  render() {
    return (
      <div>
        <h1>About</h1>
        <p>This is the about page. <Link to="/">Go Home</Link>.</p>
      </div>
    );
  }
}

class NoMatch extends React.Component {    
  render() {
    return (
      <div>
        <div>404: Nuttin here.</div>
        <Link to="/">Go Home</Link>
      </div>
    );
  }
}

export default Router;`),
	"/templates/layouts/head.js": []byte(`import React, { Component } from 'react';

class Head extends Component {
  render() {
    return <head>
      <meta charset="utf-8" />
      <meta name="viewport" content="width=device-width" />
      <title>{this.props.title}</title>
    </head>
  }
}

export default Head;`),
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
