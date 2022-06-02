<script>
    import blueprints from '../blueprints.js';
</script>

<h1>Add content of type:</h1>

<div class="blueprints">
    {#each blueprints as blueprint}
        <div class="blueprint">{blueprint.type}</div>
    {/each}
</div>

<style>
    .blueprints {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 40px;
    }
    .blueprint {
        background-color: #1c7fc7;
        color: white;
        border-radius: 6px;
        min-height: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        font-size: 1.25rem;
        cursor: pointer;
    }
</style>