<script>
    export let content;
    import { publish } from './publish.js';

    let sending = false;
    let sent = false;
    let failed = false;
    async function onSubmit() {
        const { type, filename } = content;
        const filePath = 'content/' +
            (type != 'index' ? type + '/' : '') +
            filename;

        sending = true;
        sent = false;
        failed = false;

        try {
            await publish(filePath, JSON.stringify(content.fields, undefined, '\t'));
            sending = false;
            sent = true;
        } catch (error) {
            sending = false;
            failed = true;
            throw error;
        }
    }
</script>

<form on:submit|preventDefault={onSubmit}>
    <button type="submit" disabled={sending}>Publish</button>
    {#if sending}Sending...{/if}
    {#if failed}Could not commit the changes.{/if}
    {#if sent}Changes committed.{/if}
</form>

<style>
    form {
        border-bottom: 1px solid #ccc;
        padding-top: .75rem;
        padding-bottom: .75rem;
    }
</style>