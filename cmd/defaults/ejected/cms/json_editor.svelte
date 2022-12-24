<script>
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';

    export let content;

    let formattedFields, previousFilepath;
    $: if (content.filepath !== previousFilepath) {
        formattedFields = syntaxHighlight(content.fields);
        previousFilepath = content.filepath;
    }

    const syntaxHighlight = json => {
        json = JSON.stringify(json, null, 4);
        json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|(true|false|null)|(-?[0-9]*\.?[0-9]*))/g, match => {
            let cls = 'syntax';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'key';
                } else {
                    cls = 'string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'boolean';
            } else if (/[0-9]/.test(match)) {
                cls = 'number';
            } else if (/null/.test(match)) {
                cls = 'null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
    }
</script>

<form>
    <div 
        class="json-editor"
        contenteditable=true
        on:input={e => content.fields = JSON.parse(e.target.textContent)}
        on:keydown={e => {
            if(e.key === "Tab"){
                document.execCommand('insertHTML', false, '&#32;&#32;&#32;&#32;');
                e.preventDefault()   
            }
        }}
    >
        {@html formattedFields}
    </div>
    <ButtonWrapper>
        <Button
            commitList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t')
                }
            ]}
            buttonText="Save"
            action={content.isNew ? 'create' : 'update'}
            encoding="text" />
        <Button
            commitList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t')
                }
            ]}
            buttonText="Delete"
            action={'delete'}
            encoding="text" />
    </ButtonWrapper>
</form>

<style>
    form {
        padding: 20px;
    }
    .json-editor {
        outline: 1px solid #ccc;
        background-color: white;
        font-family: monospace;
        font-size: small;
        white-space: pre-wrap;
        padding: 5px;
        margin-bottom: 20px;
    }
    .json-editor :global(.string) { color: #323232; }
    .json-editor :global(.number) { color: darkviolet; }
    .json-editor :global(.boolean) { color: darkblue; }
    .json-editor :global(.null) { color: magenta; }
    .json-editor :global(.key) { color: darkred; }
    .json-editor :global(.syntax) { color: firebrick; }
</style>