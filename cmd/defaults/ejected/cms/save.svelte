<script>
    export let content;
    import { publish } from './publish.js';
    import originalContent from '../content.js';

    let status;
    async function onSubmit() {
        const { type, filename } = content;
        const filePath = 'content/' + (type != 'index' ? type + '/' : '') + filename;
        status = "sending";
        try {
            await publish(filePath, JSON.stringify(content.fields, undefined, '\t'));
            status = "sent";
            resetStatus();
        } catch (error) {
            status = "failed";
            resetStatus();
            throw error;
        }
    }
    const resetStatus = () => {
        setTimeout(() => {
            status = "";
        }, 700);
    }

    const resetContent = () => {
        //console.log(originalContent); TODO: This is broken
        content = originalContent.find(current => current.filename == content.filename);
    }
</script>

<div class="buttons">
    <button 
        on:click|preventDefault={onSubmit}
        type="submit"
        disabled={status}
    >
    {#if status == "sending"}
        Sending...
    {:else if status == "failed"}
        Could not commit the changes.
    {:else if status == "sent"}
        Changes committed.
    {:else}
        Publish
    {/if}
    </button>
    <button
        class="reset"
        on:click|preventDefault={resetContent}
    >
        Reset
    </button>
</div>

<style>
    .buttons {
        display: flex;
        gap: 20px;
    }
    button {
        background-color: #1c7fc7;
        border: none;
        border-radius: 6px;
        color: #fff;
        cursor: pointer;
        font-weight: bold;
        line-height: 21px;
        padding: 10px;
        width: 100%;
    }
    .reset {
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
    }
</style>