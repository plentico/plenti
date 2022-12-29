<script>
    export let commitList, shadowContent, buttonText, action, encoding;
    import { publish } from './publish.js';
    import { postLocal } from './post_local.js';
    import { env } from '../env.js';

    const local = env.local ?? false;
    let status, confirmTooltip;

    const onSubmit = async () => {
        confirmTooltip = false;
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

<div class="button">
    {#if confirmTooltip}
        <div class="confirm">
            <div class="carrot"></div>
            <div>Are you sure you want to permanently remove:</div>
            {#each commitList as commitItem}
                <div class="remove-file">{commitItem.file}</div>
            {/each}
            <div class="confirm-actions">
                <button on:click|preventDefault={onSubmit}>Yes</button>
                <button on:click|preventDefault={() => confirmTooltip = false}>Cancel</button>
            </div>
        </div>
    {/if}
    <button 
        on:click|preventDefault={() => action === "delete" ? confirmTooltip = true : onSubmit}
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
</div>

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
    .button {
        position: relative;
    }
    .confirm {
        position: absolute;
        inset: auto 0 50px auto;
        background: white;
        padding: 10px;
        box-shadow: 0px 0px 6px rgb(0 0 0 / 30%);
        border-radius: 5px;
        width: 100%;
        box-sizing: border-box;
    }
    .carrot {
        position: absolute;
        left: 46%;
        bottom: -10px;
        width: 0; 
        height: 0; 
        border-left: 10px solid transparent;
        border-right: 10px solid transparent;
        border-top: 10px solid white;
    }
    .carrot:before {
        content: "";
        position: absolute;
        left: -10px;
        bottom: -1px;
        width: 0; 
        height: 0; 
        border-left: 10px solid transparent;
        border-right: 10px solid transparent;
        border-top: 10px solid gainsboro;
        z-index: -1;
    }
    .remove-file {
        color: gray;
        margin-bottom: 8px;
        padding: 8px 0;
    }
    .confirm-actions {
        display: flex;
        gap: 10px;
    }
</style>