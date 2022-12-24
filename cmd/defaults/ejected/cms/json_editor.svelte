<script>
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';

    export let content;

    const syntaxHighlight = json => {
        json = JSON.stringify(json, null, 4);
        json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|(true|false|null)|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, match => {
            let cls = 'number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'key';
                } else {
                    cls = 'string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'boolean';
            } else if (/null/.test(match)) {
                cls = 'null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
    }

    let formattedFields = syntaxHighlight(content.fields);
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
    .json-editor :global(.string) { color: darkgreen; }
    .json-editor :global(.number) { color: darkorange; }
    .json-editor :global(.boolean) { color: darkblue; }
    .json-editor :global(.null) { color: magenta; }
    .json-editor :global(.key) { color: darkred; }
</style>