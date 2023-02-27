<script>
    export let field;
</script>

<div class="list-text">
    {#each field as text, i}
        <input
            bind:value={text}
            on:keyup={() => {
                field[i] = text;
            }}
        />
    {/each}
    <button
        on:click|preventDefault={() => {
            field = [...field, ""];
        }}
    >
        Add Item
    </button>
</div>

<style>
    input {
        margin-bottom: 10px;
        padding: 5px 7px;
        width: 100%;
        box-sizing: border-box;
    }
</style>