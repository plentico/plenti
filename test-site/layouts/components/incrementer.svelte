<script>
  import { count } from '../scripts/stores.svelte';

  function increment() {
    count.update(n => n + 1);
  }
</script>

<button on:click={increment}>
  +
</button>