<script>
    import DynamicFormInput from "../dynamic_form_input.svelte";
    export let field, showMediaModal, changingMedia, localMediaList, missingRequired, parentKeys, schema;
</script>

<fieldset>
    {#each Object.entries(field) as [key, value]}
        <DynamicFormInput 
            bind:field={field[key]}
            label={key}
            bind:showMediaModal
            bind:changingMedia
            bind:localMediaList
            bind:missingRequired
            parentKeys={parentKeys === "" ? key : parentKeys + '.' + key}
            {schema}
        />
    {/each}
</fieldset>

<style>
    fieldset {
        border: 1px solid gainsboro;
        margin: 0;
    }
</style>
