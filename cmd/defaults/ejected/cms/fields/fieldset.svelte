<script>
    import DynamicFormInput from "../dynamic_form_input.svelte";
    export let field, showMedia, changingAsset, localMediaList, parentKeys, schema, compSchema;
</script>

<fieldset>
    {#each Object.entries(field) as [key, value]}
        <DynamicFormInput 
            bind:field={field[key]}
            label={key}
            bind:showMedia
            bind:changingAsset
            bind:localMediaList
            parentKeys={parentKeys === "" ? key : parentKeys + '.' + key}
            {schema}
            {compSchema}
        />
    {/each}
</fieldset>