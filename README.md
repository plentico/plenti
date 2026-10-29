<div align="center"><img src="https://plenti.co/media/perry.png" width="200" /></div>
<h1 align="center">
  Plenti
</h1>
<div align="center">Build-Time Render Engine (aka Static Site Generator) written in pure Go</div>
<div align="center">Ships with a fully integrated Git-CMS that you can host for free/cheap right with your static site</div>
<div align="center">Website: <a href="https://plenti.co">https://plenti.co</a></div>
<br />

### Requirements:exclamation:
You don't need NodeJS, Go, or any dependency other than Plenti itself.

### Installation :floppy_disk:

Homebrew:
1. Add the tap: `brew tap plentico/homebrew-plenti`
2. Install: `brew install plenti`

Snap:
1. Install: `snap install plenti`

Scoop (**Windows is not supported yet**, see [details](https://github.com/plentico/plenti/issues/45#issuecomment-668819223)):
1. Add the bucket: `scoop bucket add plenti https://github.com/plentico/scoop-plenti`
2. Install: `scoop install plenti`

Manual:
1. Download the latest [release](https://github.com/plentico/plenti/releases)
2. Move it somewhere in your `PATH` (most likely `/usr/local/bin`)

### Getting Started :rocket:
1. Create a new site: `plenti new site my-new-site`
2. Move into the folder you created: `cd my-new-site`
3. Start up the development server: `plenti serve`
4. Navigate to the site in your browser: [localhost:3000](http://localhost:3000)


### Learning the Basics 🎓
1. Documentation: https://plenti.co/docs
2. Videos: [YouTube playlist](https://www.youtube.com/watch?v=Gr3KTOnsWEM&list=PLbWvcwWtuDm0tIrvD_xHvUXHBftbHDy5T)

### Building with Pico and Pattr :hammer_and_wrench:

Plenti renders templates with [Pico](https://github.com/plentico/pico), a Go-based template engine, and uses [Pattr](https://github.com/plentico/pattr) to make the rendered pages reactive in the browser. Plenti server-renders every page at build time and includes the data needed for Pattr to hydrate it—no Node.js build step is required.

A new site has this basic structure:

```text
my-new-site/
├── content/               # JSON content nodes
│   └── pages/
│       ├── index.json
│       ├── _defaults.json # Starting values for new CMS entries
│       └── _schema.json   # CMS field definitions
├── layouts/
│   ├── content/           # One template per content type
│   │   └── pages.pico
│   └── global/
│       ├── html.pico      # Base layout (configurable)
│       └── head.pico
├── media/                 # Media managed by the CMS
├── static/                # Files copied directly to the build
└── plenti.json            # Site configuration
```

Files and folders in `content/` whose names start with `_` configure the CMS and are not rendered as content nodes.

#### Templates

Declare incoming values with `prop` in the frontmatter between `---` fences. Pico supports JavaScript expressions, conditionals, loops, components, and scoped styles:

```pico
---
prop content;
let title = content.fields.title;
---

<article>
  <h1>{title}</h1>

  {if content.fields.body}
    <div>{content.fields.body | html}</div>
  {/if}
</article>

<style>
  h1 {
    color: rebeccapurple;
  }
</style>
```

Use `{if ...}` / `{/if}` for conditions and `{for let item in items}` / `{/for}` for loops. Pattr hydrates these expressions so reactive values continue to update in the browser. See the [Pico](https://github.com/plentico/pico) and [Pattr](https://github.com/plentico/pattr) repositories for more template and directive examples.

### Default props

Plenti provides `content`, `allContent`, `env`, and `user` to every page's base layout. The optional 404 layout receives them too. Declare each prop you use, and explicitly forward props to imported or dynamically selected components. To avoid duplicating a site's entire collection into every page, the complete `allContent` collection is supplied only to content types whose **content layout** declares `prop allContent;` and to content types with pagination routes. Other pages receive an empty array. For pages using the complete collection, Plenti writes one shared `/core/allContent.json` resource that Pattr loads before hydration; server-side rendering still receives the collection synchronously.

#### `content`

`content` describes the JSON node being rendered:

```js
{
  type: "pages",
  path: "/about",
  filepath: "pages/about.json",
  filename: "about",
  fields: {
    title: "About us",
    body: "<p>Hello!</p>"
  }
}
```

The JSON fields are grouped under `content.fields`, while Plenti's routing metadata is available alongside them:

```pico
---
prop content;
---

<h1>{content.fields.title}</h1>
<p>Content type: {content.type}</p>
<p>Source file: content/{content.filepath}</p>
```

For paginated routes, `content.pager` contains the current one-based page number.

#### `allContent`

`allContent` is an array containing every content node in `content/`, using the same shape as `content`. It is useful for blogs, event listings, navigation, related content, and other aggregate pages. It always contains the complete source collection, including on paginated pages.

```pico
---
prop allContent;
let posts = allContent.filter(node => node.type === "posts");
---

{for let post in posts}
  <article>
    <h2><a href="{post.path}">{post.fields.title}</a></h2>
    <p>{post.fields.summary}</p>
  </article>
{/for}
```

You can combine normal JavaScript array operations with Pico expressions:

```pico
---
prop allContent;
let upcomingEvents = allContent
  .filter(node => node.type === "events")
  .sort((a, b) => a.fields.date.localeCompare(b.fields.date));
---

{for let event in upcomingEvents}
  <a href="{event.path}">{event.fields.title}</a>
{/for}
```

#### `env`

`env` contains sitewide build and configuration values:

- `env.local`: whether the development server is running.
- `env.baseurl`: the configured site base URL.
- `env.fingerprint`: a unique value generated for the current build.
- `env.sitevars`: custom values from `sitevars` in `plenti.json`.
- `env.routes`: configured content-type route overrides.
- `env.cms`: CMS provider and endpoint configuration.

For example, with this `plenti.json`:

```json
{
  "sitevars": {
    "siteName": "My Plenti Site",
    "social": {
      "github": "https://github.com/plentico"
    }
  }
}
```

A template can use:

```pico
---
prop env;
---

<footer>
  <p>{env.sitevars.siteName}</p>
  <a href="{env.sitevars.social.github}">GitHub</a>
  {if env.local}
    <small>Development build</small>
  {/if}
</footer>
```

#### `user`

`user` contains browser authentication state for the integrated CMS. Static rendering starts with `user.loggedIn` set to `false`; after hydration, the CMS updates it from the browser's authentication state.

```pico
---
prop user;
---

{if user.loggedIn}
  <div plenti-cms></div>
{:else}
  <button onclick="{user.login()}">Login</button>
{/if}
```

### Components and content layouts

The base layout is configured with `base_layout` in `plenti.json` and defaults to `global/html.pico`. A typical base layout forwards the current node to the template matching its content type:

```pico
---
import Head from "./head.pico";
prop content;
prop allContent;
prop env;
prop user;
---

<!DOCTYPE html>
<html lang="en">
  <Head {env} {content} {user} />
  <body>
    {if user.loggedIn}
      <div plenti-cms></div>
    {/if}
    <main>
      <='../content/{content.type}.pico' {content} {allContent} />
    </main>
  </body>
</html>
```

Components use regular PascalCase tags. The dynamic `<=` syntax above selects `layouts/content/pages.pico` when `content.type` is `pages`. Remember that each receiving template must declare forwarded props in its own frontmatter.

### Route overrides

By default, nested content types use their path below the type folder. For example, `content/posts/hello.json` has the path `/hello`. Add a route override in `plenti.json` to change this. Routes map **content type to route pattern**:

```json
{
  "routes": {
    "posts": "blog/:fields(title)"
  }
}
```

A post titled "Hello World" now has the path `blog/hello-world`. Route values are slugified before being inserted into `:fields(...)` tokens.

### Pagination

Add a `:paginate(variable)` token to a route override to render an aggregate content node across multiple pages. The variable is evaluated in that content type's Pico template and must be a positive integer.

For example, configure a `pages` node as a paginated blog index:

```json
{
  "routes": {
    "pages": "blog/:paginate(totalPages)"
  }
}
```

Then calculate `totalPages` and select the visible items in `layouts/content/pages.pico`:

```pico
---
prop content;
prop allContent;

let postsPerPage = 2;
let currentPage = content.pager;
let posts = allContent.filter(node => node.type === "posts");
let totalPages = Math.ceil(posts.length / postsPerPage);
let start = (currentPage - 1) * postsPerPage;
let visiblePosts = posts.slice(start, start + postsPerPage);
---

{for let post in visiblePosts}
  <article>
    <a href="{post.path}">{post.fields.title}</a>
  </article>
{/for}

<nav aria-label="Blog pages">
  {if currentPage > 1}
    <a href="{currentPage === 2 ? '/blog/' : '/blog/' + (currentPage - 1)}">
      Previous
    </a>
  {/if}

  <span>Page {currentPage} of {totalPages}</span>

  {if currentPage < totalPages}
    <a href="/blog/{currentPage + 1}">Next</a>
  {/if}
</nav>
```

Page one is rendered at `/blog/`, followed by `/blog/2/`, `/blog/3/`, and so on. Plenti does not create a duplicate `/blog/1/` page. `content.pager` is available only on paginated routes; `allContent` remains unpaginated so the calculation above always sees every post.

### Contributing :purple_heart:
Plenti is still new, but constantly improving. The API might change from time to time before we hit a stable `v1.0.x` release. If you find bugs or have any questions, please open a new [issue](https://github.com/plentico/plenti/issues) to let us know! If you want to help keep Plenti free from commercial interests, please consider [making a donation](https://github.com/sponsors/plentico) so we can spend more time making improvements :seedling:
