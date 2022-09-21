<script>
    export let field;
</script>

<div
    class="textarea"
    role="textbox"
    contenteditable=true
    bind:innerHTML={field}
></div>

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
</style>