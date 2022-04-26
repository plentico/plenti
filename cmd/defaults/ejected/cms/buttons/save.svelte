<script>
    export let file, contents, action;
    import { publish } from '../publish.js';

    let status;
    async function onSubmit() {
        status = "sending";
        try {
            await publish(file, contents, action);
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
</script>

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

<style>
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
</style>