<script>
    import blueprints from '../blueprints.js';
    import ButtonWrapper from './button_wrapper.svelte';

    export let showAdd, showEditor, content, filename;
    //export let filename = "";

    let selectedType;
    const setType = type => {
        selectedType = type;
    }

    let validationErrors = [];
    const validateFilename = () => {
        // Reset errors before rechecking
        validationErrors = [];

        if (filename.length == 0) {
            validationErrors = [...validationErrors, "Empty filename is not allowed"];
        }
        if (filename.indexOf(' ') >= 0) {
            validationErrors = [...validationErrors, "Spaces not allowed in filename"];
        }
        if (filename.indexOf('~') >= 0) {
            validationErrors = [...validationErrors, "No tilde (~) allowed in filename"];
        }
        if (filename.indexOf('`') >= 0) {
            validationErrors = [...validationErrors, "No backtick (`) allowed in filename"];
        }
        if (filename.indexOf('!') >= 0) {
            validationErrors = [...validationErrors, "No exclamation points (!) allowed in filename"];
        }
        if (filename.indexOf('@') >= 0) {
            validationErrors = [...validationErrors, "No at symbols (@) allowed in filename"];
        }
        if (filename.indexOf('#') >= 0) {
            validationErrors = [...validationErrors, "No pound symbols (#) allowed in filename"];
        }
        if (filename.indexOf('$') >= 0) {
            validationErrors = [...validationErrors, "No dollar signs ($) allowed in filename"];
        }
        if (filename.indexOf('%') >= 0) {
            validationErrors = [...validationErrors, "No percentage symbols (%) allowed in filename"];
        }
        if (filename.indexOf('^') >= 0) {
            validationErrors = [...validationErrors, "No carrot symbol (^) allowed in filename"];
        }
        if (filename.indexOf('&') >= 0) {
            validationErrors = [...validationErrors, "No ampersands (&) allowed in filename"];
        }
        if (filename.indexOf('*') >= 0) {
            validationErrors = [...validationErrors, "No star symbols (*) allowed in filename"];
        }
        if (filename.indexOf('(') >= 0 || filename.indexOf(')') >= 0) {
            validationErrors = [...validationErrors, "No opening or closing round brackets ( ) allowed in filename"];
        }
        if (filename.indexOf('{') >= 0 || filename.indexOf('}') >= 0) {
            validationErrors = [...validationErrors, "No opening or closing curly brackets { } allowed in filename"];
        }
        if (filename.indexOf('[') >= 0 || filename.indexOf(']') >= 0) {
            validationErrors = [...validationErrors, "No opening or closing square brackets [ ] allowed in filename"];
        }
        if (filename.indexOf('<') >= 0 || filename.indexOf('>') >= 0) {
            validationErrors = [...validationErrors, "No opening or closing angle brackets < > allowed in filename"];
        }

        // No errors, redirect to "add" page
        if (validationErrors.length === 0) {
            history.pushState(null, '', '/');
            location.hash = '#add/' + selectedType + '/' + filename;
            showAdd = false; 
            showEditor = true;
        }
    }

</script>

{#if selectedType}
    <h1>Set {selectedType} filename:</h1>
    <div class="filename">
        <span>content/{selectedType}/</span>
        <input placeholder="filename" bind:value={filename} class="{validationErrors.length > 0 ? 'error' : ''}" />
        <span>.json</span>
    </div>
    {#if validationErrors}
        <ul class="errors">
        {#each validationErrors as error}
            <li>{error}</li>
        {/each}
        </ul>
    {/if}
    <ButtonWrapper>
        <button class="button" on:click={validateFilename}>Set Filename</button>
        <button class="button" on:click={() => setType(null)}>Go back</button>
    </ButtonWrapper>
{:else}
<h1>Add content of type:</h1>

<div class="blueprints">
    {#each blueprints as blueprint}
        <button on:click={() => setType(blueprint.type)} class="blueprint">{blueprint.type}</button>
    {/each}
</div>
{/if}

<style>
    .blueprints {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 10px;
        margin-bottom: 25px;
    }
    .blueprint {
        border-radius: 6px;
        min-height: 50px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        cursor: pointer;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        background-color: transparent;
    }
    .blueprint:hover {
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
    .button {
        margin: 25px 0;
    }
    .errors {
        color: red;
    }
</style>