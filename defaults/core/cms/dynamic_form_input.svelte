<script>
    import { onMount } from 'svelte';
    import { isDate, isTime } from './date_checker.js';
    import { isMediaPath } from './media_checker.js';

    export let field, label, showMediaModal, changingMedia, localMediaList, parentKeys, schema;
    export let shadowContent = false;

    $: if (shadowContent !== false) {
        shadowContent[label] = field;
    }

    let FieldWidget;
    onMount(async () => {
        if (schema && schema.hasOwnProperty(parentKeys)) {
            FieldWidget = (await import('./fields/' + schema[parentKeys].type + '.svelte')).default;
        } else if (typeof field === "number") {
            FieldWidget = (await import('./fields/number.svelte')).default;
        } else if (typeof field === "string") {
            if (isDate(field)) {
                FieldWidget = (await import('./fields/date.svelte')).default;
            } else if (isTime(field)) {
                FieldWidget = (await import('./fields/time.svelte')).default;
            } else if (isMediaPath(field)) {
                FieldWidget = (await import('./fields/media.svelte')).default;
            } else {
                FieldWidget = (await import('./fields/text.svelte')).default;
            }
        } else if (typeof field === "boolean") {
            FieldWidget = (await import('./fields/boolean.svelte')).default;
        } else if (field.constructor === [].constructor) {
            FieldWidget = (await import('./fields/component.svelte')).default;
        } else if (field.constructor === ({}).constructor) {
            FieldWidget = (await import('./fields/fieldset.svelte')).default;
        }
    });
</script>

{#if label !== "plenti_salt"}
<div class="field {label}">
    {#if label}
        <label for="{label}">{label}</label>    
    {/if}
    {#if field === null}
        <div>field is null</div>
    {:else if field === undefined}
        <div>field is undefined</div>
    {:else if FieldWidget}
        <svelte:component
            this={FieldWidget}
            bind:field
            {label}
            bind:showMediaModal
            bind:changingMedia
            bind:localMediaList
            {parentKeys}
            {schema}
        />
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