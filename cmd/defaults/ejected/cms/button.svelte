<script>
    export let commitList, shadowContent, buttonText, action, encoding;
    import { publish } from './publish.js';
    import { postLocal } from './post_local.js';
    import { env } from '../env.js';

    const local = env.local ?? false;

    let status;
    const onSubmit = async () => {
        status = "sending";
        try {
            if (local) {
                await postLocal(commitList, shadowContent, action, encoding);
            } else {
                await publish(commitList, shadowContent, action, encoding);
            }
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
            commitList = [];
        }, 900);
    }
</script>

<button 
    on:click|preventDefault={onSubmit}
    type="submit"
    disabled={status}
    class="{status}"
>
    {#if status == "sending"}
        Sending...
    {:else if status == "failed"}
        Error saving
    {:else if status == "sent"}
        Changes committed
    {:else}
        {buttonText}
    {/if}
</button>

<style>
    button {
        color: white;
        background-color: #1c7fc7;
    }
    button.sent {
        background-color: darkgreen;
    }
    button.failed {
        background-color: darkred;
    }
</style>