<script>
  import Head from './head.svelte';
  import Nav from './nav.svelte';
  import Footer from './footer.svelte';
  import { makeTitle } from '../scripts/make_title.svelte';

  export let content, layout, allContent, allLayouts, env, user;
</script>

<html lang="en">
<Head title={makeTitle(content.filename)} {env} />
<body>
  {#if user && $user.isAuthenticated}
      <svelte:component this={$user.menu} {user} bind:content={content} />
  {/if}
  <main>
    <Nav />
    <div class="container">
      <svelte:component this={layout} {...content.fields} {content} {allContent} {allLayouts} {user} />
      <br />
    </div>
    <Footer {allContent} />
  </main>
</body>
</html>

<style>
  main {
    flex: 1 0 auto;
  }
</style>
