<script>
    export let content;
    import DynamicFormInput from './dynamic_form_input.svelte';
    import Save from './save.svelte';
</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        <div class="field">
            <label for="{label}">{label}</label>
            <DynamicFormInput bind:field={content.fields[label]} {label} />
        </div>
    {/each}
    <Save bind:content={content} />
</form>

<style>
    form {
        padding: 20px;
    }
    label {
        display: block;
    }
    .field {
        margin-bottom: 20px;
    }
</style>