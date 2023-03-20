<script>
    import {flip} from "svelte/animate";

    export let field;

    const removeItem = removeIndex => {
        field = field.filter((text, index) => index !== removeIndex);
    }
</script>

<div class="list-text">
    {#each field as text, i (i)}
        <div 
            class="list-text-item"
            animate:flip|local={{duration: 200}}
        >
            <div class="grip">
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-grip-horizontal" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                    <circle cx="5" cy="9" r="1" />
                    <circle cx="5" cy="15" r="1" />
                    <circle cx="12" cy="9" r="1" />
                    <circle cx="12" cy="15" r="1" />
                    <circle cx="19" cy="9" r="1" />
                    <circle cx="19" cy="15" r="1" />
                </svg>
            </div>
            <input
                bind:value={text}
                on:keyup={() => {
                    field[i] = text;
                }}
            />
            <button 
                class="close"
                on:click|preventDefault={() => removeItem(i)}
            >
                <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 0 24 24" width="16"><path d="M0 0h24v24H0z" fill="none"/><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
        </div>
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
    .list-text-item {
        display: flex;
        gap: 5px;
    }
    .grip, .close {
        cursor: grab;
        height: 100%;
        display: flex;
    }
    .close {
        padding: 5px;
    }
</style>