<script>
  import { count } from '../scripts/stores.svelte';

  function decrement() {
    count.update(n => n - 1);
  }
</script>

<button on:click={decrement}>
  -
</button>