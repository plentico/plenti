<script>
    export let schema, parentKeys, field;

    let disabled, hidden;
    if (schema && schema[parentKeys]?.options) {
        disabled = schema[parentKeys].options.includes("disabled");
        hidden = schema[parentKeys].options.includes("hidden");
    }
</script>

{#if disabled}
    <div
        class="textarea disabled"
        role="textbox"
        contenteditable=false
        bind:textContent={field}
    ></div>
{:else if !hidden}
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
        word-break: break-all;
    }
    .textarea.disabled {
        resize: none;
        background-color: whitesmoke;
        border: 1px solid gainsboro;
    }
</style>