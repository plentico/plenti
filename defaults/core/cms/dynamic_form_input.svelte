<script>
    import { isDate, isTime } from './date_checker.js';
    import { isMediaPath } from './media_checker.js';
    // Discoverable Widgets
    import Component from './fields/component.svelte';
    import Fieldset from './fields/fieldset.svelte';
    import Media from './fields/media.svelte';
    import Date from './fields/date.svelte';
    import Time from './fields/time.svelte';
    import Number from './fields/number.svelte';
    import Text from './fields/text.svelte';
    import Boolean from './fields/boolean.svelte';
    // Schema Defined Widgets
    import ListText from './fields/list_text.svelte';
    import Checkbox from './fields/checkbox.svelte';
    import Radio from './fields/radio.svelte';
    import Wysiwyg from './fields/wysiwyg.svelte';
    import Select from './fields/select.svelte';
    import Reference from './fields/reference.svelte';
    import References from './fields/references.svelte';
    import ID from './fields/id.svelte';

    export let field, label, showMediaModal, changingMedia, localMediaList, parentKeys, schema;
    export let shadowContent = false;

    $: if (shadowContent !== false) {
        shadowContent[label] = field;
    }

    let FieldWidget;
    (async () => {
        if (schema && schema.hasOwnProperty(parentKeys)) {
            try {
                FieldWidget = (await import('./fields/' + schema[parentKeys].type + '.svelte')).default;
            } catch (error) {
                FieldWidget = (await import('../../layouts/_fields/' + schema[parentKeys].type + '.svelte')).default;
            }
        }
    })();
</script>

{#if label !== "plenti_salt"}
    <div class="field {label}">
        {#if label}
            <label for="{label}">{label}</label>
        {/if}
        {#if field === null && !schema}
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
        {:else if typeof field === "number"}
            <Number bind:field {label} />
        {:else if typeof field === "string"}
            {#if isDate(field)}
                <Date bind:field />
            {:else if isTime(field)}
                <Time bind:field />
            {:else if isMediaPath(field)}
                <Media bind:field bind:showMediaModal bind:changingMedia bind:localMediaList />
            {:else}
                <Text bind:field />
            {/if}
        {:else if typeof field === "boolean"}
            <Boolean bind:field {label} />
        {:else if field?.constructor === [].constructor}
            <Component
                bind:field
                {label}
                bind:showMediaModal
                bind:changingMedia
                bind:localMediaList
                {parentKeys}
                {schema}
            />
        {:else if field?.constructor === ({}).constructor}
            <Fieldset
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