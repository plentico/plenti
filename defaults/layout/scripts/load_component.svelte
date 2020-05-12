<script context="module">
  export const loadComponent = component => {
    let compClassPromise = import("../components/" + component + ".svelte").then(res => res.default);
    // Fix "Unhandled promise rejection" error.
    // See: https://github.com/sveltejs/sapper/issues/487#issuecomment-529145749
    compClassPromise.catch(err => null)
    return compClassPromise;
  }
</script>
