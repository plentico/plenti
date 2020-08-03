<script>
  import Head from './head.svelte';
  import Nav from './nav.svelte';
  import Footer from './footer.svelte';
  import { makeTitle } from '../scripts/make_title.svelte';

  export let route, content, allContent;
</script>

<html lang="en">
<Head title={makeTitle(content.filename)} />
<body>
  <Nav />
  <main>
    <div class="container">
      <svelte:component this={route} {...content.fields} {allContent} />
      <br />
    </div>
  </main>
  <Footer {allContent} />
</body>
</html>

<style>
  body {
    font-family: 'Rubik', sans-serif;
    display: flex;
    flex-direction: column;
    margin: 0;
  }
  main {
    flex-grow: 1;
  }
  :global(.container) {
    max-width: 1024px;
    margin: 0 auto;
    flex-grow: 1;
    padding: 0 20px;
  }
  :global(:root) {
    --primary: rgb(34, 166, 237);
    --primary-dark: rgb(16, 92, 133);
    --accent: rgb(254, 211, 48);
    --base: rgb(245, 245, 245);
    --base-dark: rgb(17, 17, 17);
  }
  :global(main a) {
    position: relative;
    text-decoration: none;
    color: var(--base-dark);
    padding-bottom: 5px;
  }
  :global(main a:before) {
    content: "";
    width: 100%;
    height: 100%;
    background-image: linear-gradient(to top, var(--accent) 25%, rgba(0, 0, 0, 0) 40%);  
    position: absolute;
    left: 0;
    bottom: 2px;
    z-index: -1;   
    will-change: width;
    transform: rotate(-2deg);
    transform-origin: left bottom;
    transition: width .1s ease-out;
  }
  :global(main a:hover:before) {
    width: 0;
    transition-duration: .15s;
  }
</style>
