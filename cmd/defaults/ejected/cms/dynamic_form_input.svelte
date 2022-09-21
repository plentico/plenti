<script>
    import { isDate } from './dates.js';
    import { isAssetPath } from './assets_checker.js';
    import Checkbox from './fields/checkbox.svelte';
    import Radio from './fields/radio.svelte';
    import Wysiwyg from './fields/wysiwyg.svelte';
    import Component from './fields/component.svelte';
    import Fieldset from './fields/fieldset.svelte';
    import Asset from './fields/asset.svelte';
    import Select from './fields/select.svelte';
    import Autocomplete from './fields/autocomplete.svelte';
    import ID from './fields/id.svelte';
    import Date from './fields/date.svelte';
    import Text from './fields/text.svelte';
    import Boolean from './fields/boolean.svelte';
    import schemas from '../schemas.js';

    export let field, label, showMedia, changingAsset, localMediaList, parentKeys, content;

    let schema = schemas[content.type];
</script>

<div class="field {label}">
    {#if label}
        <label for="{label}">{label}</label>    
    {/if}
    {#if field === null}
        {field} is null
    {:else if field === undefined}
        {field} is undefined
    {:else if schema && schema.hasOwnProperty(parentKeys)}
        {#if schema[parentKeys].type === "checkbox"}
            <Checkbox {schema} {parentKeys} bind:field />
        {/if}
        {#if schema[parentKeys].type === "radio"}
            <Radio {schema} {parentKeys} bind:field />
        {/if}
        {#if schema[parentKeys].type === "select"}
            <Select {schema} {parentKeys} bind:field />
        {/if}
        {#if schema[parentKeys].type === "wysiwyg"}
            <Wysiwyg {schema} {parentKeys} bind:field />
        {/if}
        {#if schema[parentKeys].type === "autocomplete"}
            <Autocomplete {schema} {parentKeys} bind:field />
        {/if}
        {#if schema[parentKeys].type === "id"}
            <ID bind:field />
        {/if}
    {:else if typeof field === "number"}
        <input id="{label}" type="number" bind:value={field} />
    {:else if typeof field === "string"}
        {#if isDate(field)}
            <Date bind:field />
        {:else if isAssetPath(field)}
            <Asset bind:field bind:showMedia bind:changingAsset bind:localMediaList />
        {:else}
            <Text bind:field />
        {/if}
    {:else if typeof field === "boolean"}
        <Boolean bind:field {label} />
    {:else if field.constructor === [].constructor}
        <Component bind:field bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {content} />
    {:else if field.constructor === ({}).constructor}
        <Fieldset bind:field bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {content} />
    {/if}
</div>

<style>
    label {
        display: block;
    }
    .field {
        margin-bottom: 20px;
    }
    .field:last-of-type {
        margin-bottom: 0;
    }
</style>