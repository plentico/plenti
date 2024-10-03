<script>
    import defaults from '../../generated/defaults.js';
    import ButtonWrapper from './button_wrapper.svelte';
    import validateFilename from './validate_filename.js';

    export let showContentModal, showAdd, showEditor, env, selectedType;
    let filename = "";
    let baseurl = env.baseurl ? env.baseurl : '/';

    const setType = type => {
        selectedType = type;
    }

    let validationErrors = [];
    const checkFilename = () => {
        validationErrors = validateFilename(filename, selectedType);
        // No errors, redirect to "add" page
        if (validationErrors.length === 0) {
            showContentModal = false;
            redirectAndEdit(baseurl + '#add/' + selectedType + '/' + filename);
        }
    }

    const redirectAndEdit = path => {
        history.pushState(null, '', path);
        showAdd = false; 
        showEditor = true;
    }

</script>

{#if selectedType}
    <h1>Add {selectedType}:</h1>
    <div class="filename">
        <span>content/{selectedType}/</span>
        <input placeholder="filename" autofocus bind:value={filename} class="{validationErrors.length > 0 ? 'error' : ''}" />
        <span>.json</span>
    </div>
    {#if validationErrors}
        <ul class="errors">
        {#each validationErrors as error}
            {#if typeof error === "object"}    
                <li>{error.message} <span class="error-link" on:click={() => redirectAndEdit(error.contentPath)}>Edit Content</span>?</li>
            {:else}
                <li>{error}</li>
            {/if}
        {/each}
        </ul>
    {/if}
    <ButtonWrapper>
        <button on:click={checkFilename} class="primary">Set Filename</button>
        <button on:click={() => {
            showAdd = false;
            setType("");
        }}>Go back</button>
    </ButtonWrapper>
{:else}
    <h1>Add content of type:</h1>
    <div class="defaults">
        {#each defaults as defaultContent}
            <ButtonWrapper>
                <button 
                    on:click={() => setType(defaultContent.type)}
                    class="default"
                >
                    {defaultContent.type}
                </button>
            </ButtonWrapper>
        {/each}
    </div>
{/if}

<style>
    .defaults {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 10px;
        margin-bottom: 25px;
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
    .default {
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
    }
    .primary,
    .default:hover {
        background-color: #1c7fc7;
        color: white;
    }
    .filename input {
        background: #ededed;
        border: none;
        border-bottom: 3px solid;
        line-height: 2rem;
        font-size: 1.5rem;
        padding: 0 5px;
        width: 55%;
    }
    input.error {
        background-color: #ffc0c0;
    }
    .errors {
        color: red;
    }
    .error-link {
        cursor: pointer;
        color: #1c7fc7;
        text-decoration: underline;
    }
</style>
