<script>
    import blueprints from '../blueprints.js';
    import ButtonWrapper from './button_wrapper.svelte';

    let selectedType;
    const setType = type => {
        selectedType = type;
    }

</script>

{#if selectedType}
    <h1>Set filename for content:</h1>
    <div class="filename">
        <span>content/{selectedType}/</span>
        <input placeholder="filename" />
        <span>.json</span>
    </div>
    <ButtonWrapper>
        <button class="button">Save</button>
        <button class="button" on:click={() => setType(null)}>Go back</button>
    </ButtonWrapper>
{:else}
<h1>Add content of type:</h1>

<div class="blueprints">
    {#each blueprints as blueprint}
        <button on:click={() => setType(blueprint.type)} class="blueprint">{blueprint.type}</button>
    {/each}
</div>
{/if}

<style>
    .blueprints {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 10px;
        margin-bottom: 25px;
    }
    .blueprint {
        border-radius: 6px;
        min-height: 50px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        cursor: pointer;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        background-color: transparent;
    }
    .blueprint:hover {
        background-color: #1c7fc7;
        color: white;
    }
    .filename input {
        background: #ededed;
        border: none;
        border-bottom: 3px solid;
        line-height: 2rem;
        font-size: 1.5rem;
        padding: 0 5px;
        width: 55%;
    }
    .button {
        margin: 25px 0;
    }
</style>