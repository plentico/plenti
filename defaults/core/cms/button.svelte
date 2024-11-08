<script>
    import { commitGitlab } from './providers/gitlab.js';
    import { commitGitea } from './providers/gitea.js';
    import { postLocal } from './providers/local.js';
    import { env } from '../../generated/env.js';
    import { findFileReferences } from './file_references.js';

    export let commitList, shadowContent, buttonText, action, encoding, user, afterSubmit, status;
    export let buttonStyle = "primary";
    const local = env.local ?? false;
    const provider = env.cms.provider.toLowerCase();

    let confirmTooltip;
    const onSubmit = async () => {
        confirmTooltip = false;
        status = "sending";
        try {
            if (local) {
                await postLocal(commitList, shadowContent, action, encoding, user);
            } else if (!provider || provider === "gitlab") {
                await commitGitlab(commitList, shadowContent, action, encoding, user);
            } else if (provider === "gitea" || provider === "forgejo") {
                await commitGitea(commitList, shadowContent, action, encoding, user);
            }
            status = "sent";
            afterSubmit?.();
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
            <div class="warning">Are you sure you want to permanently remove:</div>
            {#each commitList as commitItem}
                <div class="delete-filepath">
                    {commitItem.file}
                    {#if findFileReferences(commitItem.file).length > 0}
                        <div class="file-reference">
                            This file is being used on:
                        </div>
                        {#each findFileReferences(commitItem.file) as reference, i}
                            <a href="{reference}">{reference}</a>{#if i < findFileReferences(commitItem.file).length -1},&nbsp;{/if}
                        {/each}
                    {/if}
                </div>
            {/each}
            <div class="confirm-actions">
                <button
                    on:click|preventDefault={onSubmit}
                    class="primary"
                >
                    Yes
                </button>
                <button
                    on:click|preventDefault={() => confirmTooltip = false}
                    class="secondary"
                >
                    Cancel
                </button>
            </div>
        </div>
    {/if}
    <button 
        on:click|preventDefault={() => action === "delete" ? confirmTooltip = true : action ? onSubmit() : null}
        on:click
        type="submit"
        disabled={status}
        class="{status} {buttonStyle}"
    >
        {#if status == "sending"}
            Sending...
        {:else if status == "failed"}
            Error saving
        {:else if status == "sent"}
            Changes committed
        {:else if status == "missing_required"}
            Missing Required
        {:else}
            {buttonText}
        {/if}
    </button>
</div>

<style>
    .button {
        position: relative;
        width: 100%;
    }
    button.primary {
        color: white;
        background-color: #1c7fc7;
    }
    button.sent {
        background-color: darkgreen;
    }
    button.failed {
        background-color: darkred;
    }
    button[disabled] {
        background: gray;
        cursor: not-allowed;
    }
    button {
        width: 100%;
        border-radius: 6px;
        cursor: pointer;
        border: none;
        font-weight: bold;
        line-height: 21px;
        padding: 10px;
    }
    .secondary {
        background-color: #e7e7e7;
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
    .warning {
        margin-bottom: 5px;
    }
    .delete-filepath {
        color: gray;
        word-wrap: break-word;
        word-break: break-word;
        margin-bottom: 10px;
    }
    .file-reference {
        color: darkred;
    }
    .confirm-actions {
        display: flex;
        gap: 10px;
    }
</style>
