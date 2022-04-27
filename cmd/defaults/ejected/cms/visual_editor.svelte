<script>
    export let content;
    import DynamicFormInput from './dynamic_form_input.svelte';
    import Buttons from './buttons/buttons.svelte';
    import Save from './buttons/save.svelte';
</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        <div class="field">
            <label for="{label}">{label}</label>
            <DynamicFormInput bind:field={content.fields[label]} {label} />
        </div>
    {/each}
    <Buttons>
        <Save mediaList={[{file: content.filepath, contents: JSON.stringify(content.fields, undefined, '\t')}]} action="update" encoding="text" />
        <button>Reset</button>
    </Buttons>
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