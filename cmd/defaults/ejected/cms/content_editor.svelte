<script>
    import { onMount } from 'svelte';

    export let content;

    const loaded = Promise.resolve()
        .then(() => import('https://unpkg.com/codemirror@5.65.1/lib/codemirror.js'))
        .then(() => import('https://unpkg.com/codemirror@5.65.1/mode/javascript/javascript.js'));
    let container;
    let editor;
    onMount(async () => {
        await loaded;
        editor = new CodeMirror(container, {
            mode: 'javascript',
        });
        editor.on('change', () => {
            try {
                content.fields = JSON.parse(editor.getValue());
            } catch (error) {
                if (!(error instanceof SyntaxError)) {
                    throw error;
                }
            }
        });
    });
    $: if (editor && !editor.hasFocus()) {
        editor.setValue(JSON.stringify(content.fields, undefined, 4));
    }
</script>

<style>
    form {
        border-bottom: 1px solid #ccc;
        padding-top: .75rem;
        padding-bottom: .75rem;
    }

    .editor-container {
        border: 1px solid #ccc;
        margin-bottom: .75rem;
    }

    .editor-container :global(.CodeMirror) {
        height: auto;
        max-height: 250px;
    }

</style>

<svelte:head>
    <link rel="stylesheet" href="https://unpkg.com/codemirror@5.65.1/lib/codemirror.css">
</svelte:head>

<form on:submit|preventDefault>
    <div class="editor-container" bind:this={container}></div>
</form>