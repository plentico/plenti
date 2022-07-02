<script>
    export let content, showMedia, changingAsset;
    import DynamicFormInput from './dynamic_form_input.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';
</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        <div class="field">
            <label for="{label}">{label}</label>
            <DynamicFormInput bind:field={content.fields[label]} {label} bind:showMedia bind:changingAsset />
        </div>
    {/each}
    <ButtonWrapper>
        <Button
            commitList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t'),
                },
            ]}
            buttonText="Save"
            action={content.isNew ? 'create' : 'update'}
            encoding="text" />
        <Button
            commitList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t'),
                },
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
    label {
        display: block;
    }
    .field {
        margin-bottom: 20px;
    }
</style>