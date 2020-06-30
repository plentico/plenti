<div align="center"><img src="https://plenti.co/assets/perry.png" width="200" /></div>
<h1 align="center">
  Plenti
</h1>
<div align="center">Static Site Generator with Go backend and Svelte frontend</div>
<div align="center">Website: <a href="https://plenti.co">https://plenti.co</a></div>
<br />

### Requirements:exclamation:
You must have [NodeJS](https://nodejs.org/) version 13 or newer

### Installation :floppy_disk:

Homebrew:
1. Add the tap: `brew tap plentico/homebrew-plenti`
2. Install: `brew install plenti`

Snap:
1. Install: `snap install plenti` (TODO: need to fix https://github.com/plentico/plenti/issues/31 before this will work properly)

Scoop:
1. Add the bucket: `scoop bucket add org https://github.com/plentico/scoop-plenti.git`
2. Install: `scoop install plentico/scoop-plenti`

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

<details>
<summary>Types</summary>

### Types
The `content/` folder in a project is where all your data lives (in JSON format). This is typically divided into multiple subfolders that define your _types_. Types are just a way to group content of a similar structure. Individual files inside a type are very flexible, in fact you can define any field schema you'd like and there are no required keys. Even though files may be grouped together as a type, they can actually have variability between them in terms of their field structure - just make sure you account for this in your corresponding `layout/content/` files!

**Single file types**: Anything that appears at the first level within the content folder is a type. This can include single files such as `index.json` and `404.json`, which are also types, but only have a one-off data source. You can define your own single file types this way if you'd like.

**Blueprints**: There is an optional, special named file that goes inside your individual type folders named `_blueprint.json`. This defines the default field schema for that specific type. The _keys_ of the blueprint correspond to field names used in the content files and the _values_ tell the kind of field that is being used. **TODO**: Currently the blueprint doesn't do much and there is no list of standardized _values_, but in the future this will be fleshed out and it will aid in generating scaffoling and tying into the cms (see https://github.com/plentico/plenti/issues/15).

**Paths**: The endpoint nodes for your pages (of whatever Type) will be defined by your data source. By default this corresponds to the structure of folders and files in your `content/` folder, for example:

- `content/index.json` = `https://example.com/`
- `content/blog/post1.json` = `https://example.com/blog/post1`
- `content/events/event1.json` = `https://example.com/events/event1`

You can overide the default path structure in the site's configuration file (`plenti.json`). For example if you had a Type called `pages` and you wanted it to appear at the top level of the site and not in the format `https://example.com/pages/page1`, you could add the following to `plenti.json`:

```json
"types": {
  "pages": "/:filename"
}
```
This would allow a content file located at `content/pages/page1.json` to appear in the following format: `https://example.com/page1`. 

You can use any custom key that you define in your content source, e.g. `:title`, `:date`, etc in your path, for example:
```json
"types": {
  "blog": "/blog/:field(author)/:field(title)"
}
```

If you want to have a content source without a path (no node endpoint that site visitors can access), simply delete the corresponding svelte template in `layout/content/`. You can do this automatically use the "endpoint" flag when creating a new type, for example: `plenti new type YOUR_TYPE --endpoint=false`

</details>

<details>
<summary>Layout</summary>

### Layout
All the templating is done in "disappearing" JS component framework called [Svelte](https://svelte.dev/). Svelte offers a simplified syntax and creates a welcoming developer experience for folks coming directly from an HTML/CSS background. It also offers some performance benefits over similar frameworks since it doesn't require a virtual dom and its runtime is rather small.

**layout/global/html.svelte**: This file is important and changing its name will break your app. You could also potentially break your routing if you're not careful with `<svelte:component this={route} {...node.fields} {allNodes} />`. Once you're aware of those two things, this file shouldn't be too scary and is meant for you to customize.

**layout/content/**: Files that live in this folder correspond directly to the Types defined in your content source. For example if you have blog Type (`content/blog/post-whatever.json`) you would create a corresponding template at `layout/content/blog.svelte`. One template should be used per Type and it will feed many content files to create individual nodes (endpoints).

The rest of the structure is really up to you. We try to create logical default folders, such as `layout/components/`for reusable widgets and `layout/scripts/` for helper functions, but feel free to completely change these and make the structure your own.


</details>

### Contributing :purple_heart:
Plenti is brand new and needs to be test driven a bit to work out the kinks. If you find bugs or have any questions, please open a new [issue](https://github.com/plentico/plenti/issues) to let us know! Thank you for being patient while Plenti grows :seedling:
