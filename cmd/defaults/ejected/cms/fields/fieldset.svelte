<script>
    import DynamicFormInput from "../dynamic_form_input.svelte";
    export let field, showMedia, changingAsset, localMediaList, parentKeys, schema;
</script>

<fieldset>
    {#each Object.entries(field) as [key, value]}
        <div class="field">
            <label for={key}>{key}</label>
            <DynamicFormInput bind:field={field[key]} label={key} bind:showMedia bind:changingAsset bind:localMediaList parentKeys={parentKeys + '.' + key} {schema} />
        </div>
    {/each}
</fieldset>