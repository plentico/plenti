<script>
    export let content, showMediaModal, changingMedia, localMediaList, shadowContent, user;
    import DynamicFormInput from './dynamic_form_input.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';
    import schemas from '../../generated/schemas.js';

    $: schema = schemas[content.type];

</script>

<form>
    {#key content.filepath}
        {#each Object.entries(content.fields) as [label, field]}
            {#if schema}
                {#each Object.entries(schema) as schema_field}
                    {#if schema_field[1]?.before === label}
                        <DynamicFormInput 
                            field={shadowContent[schema_field[0]]}
                            bind:shadowContent
                            label={schema_field[0]}
                            bind:showMediaModal
                            bind:changingMedia
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
                bind:showMediaModal
                bind:changingMedia
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
                            bind:showMediaModal
                            bind:changingMedia
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
                {user}
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
                buttonStyle="secondary"
                action={'delete'}
                encoding="text"
                {user}
            />
        </ButtonWrapper>
    {/key}
</form>

<style>
    form {
        padding: 20px;
    }
</style>