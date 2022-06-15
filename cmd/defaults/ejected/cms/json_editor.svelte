<script>
    import { onMount } from 'svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';

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

<svelte:head>
    <link rel="stylesheet" href="https://unpkg.com/codemirror@5.65.1/lib/codemirror.css">
</svelte:head>

<form>
    <div class="editor-container" bind:this={container}></div>
    <ButtonWrapper>
        <Button
            mediaList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t')
                }
            ]}
            buttonText="Publish"
            action={content.isNew ? 'create' : 'update'}
            encoding="text" />
        <button>Reset</button>
    </ButtonWrapper>
</form>

<style>
    form {
        padding: 20px;
    }
    .editor-container {
        border: 1px solid #ccc;
        margin-bottom: .75rem;
    }
    .editor-container :global(.CodeMirror) {
        height: auto;
    }
</style>