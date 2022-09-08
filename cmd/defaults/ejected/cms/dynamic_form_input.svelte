<script>
    import { isDate, makeDate, formatDate } from './dates.js';
    import { isAssetPath } from './assets_checker.js';
    import Checkbox from './fields/checkbox.svelte';
    import Wysiwyg from './fields/wysiwyg.svelte';
    import Component from './fields/component.svelte';
    import Group from './fields/group.svelte';
    import Asset from './fields/asset.svelte';

    export let field, label, showMedia, changingAsset, localMediaList, parentKeys, schema;

    const bindDate = date => {
        field = formatDate(date, field);
    }
</script>

<div class="field">
    {#if field === null}
        {field} is null
    {:else if field === undefined}
        {field} is undefined
    {:else if schema && schema.hasOwnProperty(parentKeys)}
        {#if schema[parentKeys].type === "checkbox"}
            <Checkbox {schema} {parentKeys} bind:field />
        {/if}
        {#if schema[parentKeys].type === "wysiwyg"}
            <Wysiwyg {schema} {parentKeys} bind:field />
        {/if}
    {:else if typeof field === "number"}
        <input id="{label}" type="number" bind:value={field} />
    {:else if typeof field === "string"}
        {#if isDate(field)}
            <input type="date" value={makeDate(field)} on:input={date => bindDate(date.target.value)} />
        {:else if isAssetPath(field)}
            <Asset bind:field bind:showMedia bind:changingAsset bind:localMediaList />
        {:else}
            <div
                class="textarea"
                role="textbox"
                contenteditable=true
                bind:innerHTML={field}
            ></div>
        {/if}
    {:else if typeof field === "boolean"}
        <label><input id="{label}" type="checkbox" bind:checked={field} /> {field}</label>
    {:else if field.constructor === [].constructor}
        <Component bind:field bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {schema} />
    {:else if field.constructor === ({}).constructor}
        <Group bind:field bind:showMedia bind:changingAsset bind:localMediaList bind:parentKeys {schema} />
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
    .textarea {
        background: white;
        border: 1px solid gainsboro;
        resize: vertical;
        overflow: auto;
        padding: 7px;
        font-family: sans-serif;
        font-size: small;
        white-space: pre-wrap;
    }
    /*
    textarea {
        width: 100%;
        resize: vertical;
        box-sizing: border-box;
        padding: 7px;
        border: 1px solid gainsboro;
    }
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
        width: 100%;
        box-sizing: border-box;
    }
    */
</style>