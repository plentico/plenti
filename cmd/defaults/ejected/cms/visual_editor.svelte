<script>
    export let content, showMedia, changingAsset, localMediaList, shadowContent;
    import DynamicFormInput from './dynamic_form_input.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';
    import schemas from '../schemas.js';

    $: schema = schemas[content.type];

</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        {#if schema}
            {#each Object.entries(schema) as schema_field}
                {#if schema_field[1]?.before === label}
                    <DynamicFormInput 
                        field={shadowContent[schema_field[0]]}
                        bind:shadowContent
                        label={schema_field[0]}
                        bind:showMedia
                        bind:changingAsset
                        bind:localMediaList
                        parentKeys={schema_field[0]}
                        {schema}
                    />
                {/if}
            {/each}
        {/if}
        <DynamicFormInput 
            bind:field={content.fields[label]}
            {label}
            bind:showMedia
            bind:changingAsset
            bind:localMediaList
            parentKeys={label}
            {schema}
        />
        {#if schema}
            {#each Object.entries(schema) as schema_field}
                {#if schema_field[1]?.after === label}
                    <DynamicFormInput 
                        field={shadowContent[schema_field[0]]}
                        bind:shadowContent
                        label={schema_field[0]}
                        bind:showMedia
                        bind:changingAsset
                        bind:localMediaList
                        parentKeys={schema_field[0]}
                        {schema}
                    />
                {/if}
            {/each}
        {/if}
    {/each}
    <ButtonWrapper>
        <Button
            commitList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t'),
                },
            ]}
            {shadowContent}
            buttonText="Save"
            action={content.isNew ? 'create' : 'update'}
            encoding="text"
        />
        <Button
            commitList={[
                {
                    file: content.filepath,
                    contents: JSON.stringify(content.fields, undefined, '\t'),
                },
            ]}
            {shadowContent}
            buttonText="Delete"
            action={'delete'}
            encoding="text"
        />
    </ButtonWrapper>
</form>

<style>
    form {
        padding: 20px;
    }
</style>