<script>
    export let schema, parentKeys, field;

    let disabled;
    if (schema && schema[parentKeys]?.options) {
        disabled = schema[parentKeys].options.includes("disabled");
    }
</script>

{#if disabled}
    <div
        class="textarea disabled"
        role="textbox"
        contenteditable=false
        bind:textContent={field}
    ></div>
{:else}
    <div
        class="textarea"
        role="textbox"
        contenteditable=true
        bind:textContent={field}
    ></div>
{/if}

<style>
    .textarea {
        background: white;
        border: 1px solid gainsboro;
        resize: vertical;
        overflow: auto;
        padding: 7px;
        font-family: sans-serif;
        font-size: small;
        white-space: pre-wrap;
    }
    .textarea.disabled {
        resize: none;
        background-color: whitesmoke;
        border: 1px solid gainsboro;
    }
</style>