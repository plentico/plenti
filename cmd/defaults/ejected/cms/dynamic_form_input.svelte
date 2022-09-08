<script>
    import { isDate, makeDate, formatDate } from './dates.js';
    import { isAssetPath, isImagePath, isDocPath } from './assets_checker.js';
    import Checkbox from './fields/checkbox.svelte';
    import Wysiwyg from './fields/wysiwyg.svelte';
    import Component from './fields/component.svelte';
    import Group from './fields/group.svelte';

    export let field, label, showMedia, changingAsset, localMediaList, parentKeys, schema;

    const bindDate = date => {
        field = formatDate(date, field);
    }

    let originalAsset;
    const swapAsset = () => {
        originalAsset = field;
        changingAsset = field;
        showMedia = true;
    }
    $: if (changingAsset) {
        if (field === originalAsset) {
            field = changingAsset;
        }
    }

    // If an img path is 404, load the data image instead
    const loadDataImage = imgEl => {
        // Get src from img that was clicked on in visual editor
        let src = imgEl.target.attributes.src.nodeValue;
        // Load all image on the page with that source
        // TODO: Could load images not related to this field specifically
        let allImg = document.querySelectorAll('img[src="' + src + '"]');
        allImg.forEach(i => {
            localMediaList.forEach(asset => {
                // Check if the field path matches a recently uploaded file in memory
                if(asset.file === field) {
                    // Set the source to the data image instead of the path that can't be found
                    i.src = asset.contents; 
                }
            });
        });
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
    {#if false}
        {console.log(parentKeys.split('.').reduce((o,i)=> o[i], schema))}
        <Wysiwyg bind:field />
    {/if}
{:else if typeof field === "number"}
    <input id="{label}" type="number" bind:value={field} />
{:else if typeof field === "string"}
    {#if isDate(field)}
        <input type="date" value={makeDate(field)} on:input={date => bindDate(date.target.value)} />
    {:else if isAssetPath(field)}
        <div class="thumbnail-wrapper">
            {#if isImagePath(field)}
                <img src="{field}" alt="click to change thumbnail" class="thumbnail" on:error={imgEl => loadDataImage(imgEl)} />
            {:else if isDocPath(field)}
                <embed src="{field}" class="thumbnail" />
            {/if}
            <button class="swap" on:click|preventDefault={swapAsset}>Change Asset</button>
        </div>
    {:else if field.length < 50}
        <input id="{label}" type="text" bind:value={field} />
    {:else}
        <Wysiwyg bind:field />
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
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
        width: 100%;
        box-sizing: border-box;
    }
    .thumbnail-wrapper {
        height: 115px;
        overflow: hidden;
        position: relative;
    }
    .thumbnail {
        max-width: 200px;
    }
    button.swap {
        cursor: pointer;
        position: absolute;
        top: 0;
        left: 0;
        width: 200px;
        height: 115px;
        border: 0;
        background-color: transparent;
        color: transparent;
        font-size: 1.25rem;
        transition: all .15s;
    }
    button.swap:hover {
        background-color: rgba(0, 0, 0, .75);
        color: white;
    }
</style>