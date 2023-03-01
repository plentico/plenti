<div align="center"><img src="https://plenti.co/media/perry.png" width="200" /></div>
<h1 align="center">
  Plenti
</h1>
<div align="center">Build-Time Render Engine (aka Static Site Generator) with Go backend and Svelte frontend</div>
<div align="center">Ships with a fully integrated Git-CMS that you can host for cheap/free right with your static site</div>
<div align="center">Website: <a href="https://plenti.co">https://plenti.co</a></div>
<br />

### Requirements:exclamation:
~~You must have [NodeJS](https://nodejs.org/) version 13 or newer~~ As of `v0.2.0` you no longer need NodeJS, Go, or any dependency other than Plenti itself.

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

### Contributing :purple_heart:
Plenti is still new, but constantly improving. The API might change from time to time before we hit a stable `v1.0.x` release. If you find bugs or have any questions, please open a new [issue](https://github.com/plentico/plenti/issues) to let us know! If you want to help keep Plenti free from commercial interests, please consider [making a donation](https://github.com/sponsors/plentico) so we can spend more time making improvements :seedling:
