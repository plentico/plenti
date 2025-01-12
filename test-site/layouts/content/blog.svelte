<script>
	export let title, body, author, date, store, source, content;
  import Uses from "../components/source.svelte";

  // Svelte store example:
  import { count } from '../scripts/stores.svelte';
  import Incrementer from '../components/incrementer.svelte';
  import Decrementer from '../components/decrementer.svelte';
  let count_value;
  const unsubscribe = count.subscribe(value => {
    count_value = value;
  });

  // Content driven dynamic components example:
  export let components, allLayouts;
</script>

<h1>{title}</h1>

<p><em>{#if author}Written by {author}{/if}{#if date}&nbsp;on {date}{/if}</em></p>

<p>{@html body}</p>

{#if store}
  <h3>The count is {count_value}</h3>
  <Incrementer/>
  <Decrementer/>  
{/if}

{#if components}
	{#each components as { name }}
		<svelte:component this="{allLayouts["layouts_components_" + name + "_svelte"]}" />
	{/each}
{/if}

{#if source}
  <Uses {content} {source} />
{/if}

<p><a href=".">Back home</a></p>
