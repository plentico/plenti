<script>
    import { isDate } from './date_checker.js';
    import { isAssetPath } from './asset_checker.js';
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
    import Time from './fields/time.svelte';
    import Number from './fields/number.svelte';
    import Text from './fields/text.svelte';
    import Boolean from './fields/boolean.svelte';

    export let field, label, showMedia, changingAsset, localMediaList, parentKeys, schema, compSchema;
    export let shadowContent = false;

    $: if (shadowContent !== false) {
        shadowContent[label] = field;
    }
</script>

{#if label !== "plenti_salt"}
<div class="field {label}">
    {#if label}
        <label for="{label}">{label}</label>    
    {/if}
    {#if compSchema && compSchema.hasOwnProperty(parentKeys)}
        {#if compSchema[parentKeys].type === "component"}
            <Component bind:field {label} bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {schema} />
        {/if}
        {#if compSchema[parentKeys].type === "checkbox"}
            <Checkbox schema={compSchema} {parentKeys} bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "radio"}
            <Radio schema={compSchema} {parentKeys} bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "select"}
            <Select schema={compSchema} {parentKeys} bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "wysiwyg"}
            <Wysiwyg schema={compSchema} {parentKeys} bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "autocomplete"}
            <Autocomplete schema={compSchema} {parentKeys} bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "id"}
            <ID bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "text"}
            <Text bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "number"}
            <Number bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "boolean"}
            <Boolean bind:field {label} />
        {/if}
        {#if compSchema[parentKeys].type === "date"}
            <Date bind:field />
        {/if}
        {#if compSchema[parentKeys].type === "time"}
            <Time bind:field schema={compSchema} {parentKeys} />
        {/if}
        {#if compSchema[parentKeys].type === "asset"}
            <Asset bind:field bind:showMedia bind:changingAsset bind:localMediaList />
        {/if}
    {:else if schema && schema.hasOwnProperty(parentKeys)}
        {#if schema[parentKeys].type === "component"}
            <Component bind:field {label} bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {schema} />
        {/if}
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
        {#if schema[parentKeys].type === "text"}
            <Text bind:field />
        {/if}
        {#if schema[parentKeys].type === "number"}
            <Number bind:field />
        {/if}
        {#if schema[parentKeys].type === "boolean"}
            <Boolean bind:field {label} />
        {/if}
        {#if schema[parentKeys].type === "date"}
            <Date bind:field />
        {/if}
        {#if schema[parentKeys].type === "time"}
            <Time bind:field {schema} {parentKeys} />
        {/if}
        {#if schema[parentKeys].type === "asset"}
            <Asset bind:field bind:showMedia bind:changingAsset bind:localMediaList />
        {/if}
    {:else if typeof field === "number"}
        <Number bind:field {label} />
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
        <Component bind:field bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {schema} />
    {:else if field.constructor === ({}).constructor}
        <Fieldset bind:field bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {schema} {compSchema} />
    {:else if field === null}
        <div>field is null</div>
    {:else if field === undefined}
        <div>field is undefined</div>
    {/if}
</div> 
{/if}

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