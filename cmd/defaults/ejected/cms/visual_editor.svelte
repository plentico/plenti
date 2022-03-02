<script>
    export let content;
    import DynamicFormInput from './dynamic_form_input.svelte';
</script>

<form>
    {#each Object.entries(content.fields) as [label, field]}
        <div class="field">
            <label for="{label}">{label}</label>
            <DynamicFormInput {field} {label} bind:content={content} />
        </div>
    {/each}
</form>

<style>
    form {
        padding: 20px;
    }
    label {
        display: block;
    }
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
    }
    .field {
        margin-bottom: 20px;
    }
</style>