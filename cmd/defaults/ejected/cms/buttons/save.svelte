<script>
    export let mediaList, action, encoding;
    import { publish } from '../publish.js';

    let status;
    async function onSubmit() {
        status = "sending";
        mediaList.forEach(async mediaItem => { 
            try {
                await publish(mediaItem.file, mediaItem.contents, action, encoding);
                status = "sent";
                resetStatus();
            } catch (error) {
                status = "failed";
                resetStatus();
                throw error;
            }
        });
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