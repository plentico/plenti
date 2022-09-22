<script>
    export let content, showMedia, changingAsset, localMediaList;
    import DynamicFormInput from './dynamic_form_input.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';
    import schemas from '../schemas.js';

    let schema = schemas[content.type];
</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        <DynamicFormInput bind:field={content.fields[label]} {label} bind:showMedia bind:changingAsset bind:localMediaList parentKeys={label} {schema} />
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