<script>
    import { isDate, makeDate, formatDate } from './dates.js';
    import { isAssetPath, isImagePath, isDocPath } from './assets_checker.js';
    import Checkbox from './fields/checkbox.svelte';
    export let field, label, showMedia, changingAsset, localMediaList, parentKeys, schema;

    const bindDate = date => {
        field = formatDate(date, field);
    }

    // Accordion
    import {slide} from "svelte/transition";
    let isOpen = false;
    let openKeys = [];
    const accordion = newKey => {
        if (openKeys.length === 1 && openKeys.includes(newKey)) {
            setTimeout(() => {
                isOpen = false;
            }, 300);
        }
        if (openKeys.includes(newKey)) {
            // Remove key
            openKeys = openKeys.filter(key => key !== newKey);
        } else {
            // Add key
            openKeys = [...openKeys, newKey];
            isOpen = true;
        }
    }

    // Drag and drop
    import {flip} from "svelte/animate";
    export let removesItems = true;
    let compID;
    let ghost;
    let grabbed;
    let lastTarget;
    let mouseY = 0; // pointer y coordinate within client
    let offsetY = 0; // y distance from top of grabbed element to pointer
    let layerY = 0; // distance from top of list to top of client
    const grab = (clientY, element) => {
        // modify grabbed element
        grabbed = element;
        grabbed.dataset.grabY = clientY;
        // modify ghost element (which is actually dragged)
        ghost.innerHTML = grabbed.innerHTML;
        // record offset from cursor to top of element
        // (used for positioning ghost)
        offsetY = grabbed.getBoundingClientRect().y - clientY;
        drag(clientY);
    }
    // drag handler updates cursor position
    const drag = clientY => {
        if (grabbed) {
            mouseY = clientY;
            layerY = ghost.parentNode.getBoundingClientRect().y;
        }
    }
    // touchEnter handler emulates the mouseenter event for touch input
    const touchEnter = ev => {       
        drag(ev.clientY);
        // trigger dragEnter the first time the cursor moves over a list item
        let target = document.elementFromPoint(ev.clientX, ev.clientY).closest(".item-wrapper");
        if (target && target != lastTarget) {
            lastTarget = target;
            dragEnter(ev, target);
        }
    }
    const dragEnter = (ev, target) => {
        // swap items
        if (grabbed && target != grabbed && target.classList.contains("item-wrapper")) {
            moveItem(parseInt(grabbed.dataset.index), parseInt(target.dataset.index));
        }
    }
    // does the actual moving of items
    const moveItem = (from, to) => {
        let temp = field[from];
        field[from] = field[to];
        field[to] = temp;
    }
    const release = ev => {
        grabbed = null;
    }
    const removeItem = val => {
        field = field.filter(i => i !== val);
    }

    let textarea;

    let linkURL, linkText, linkOptions;
    const createLink = () => {
        linkURL = prompt('Enter a URL:', 'http://');
        let selectedText = document.getSelection().toString();
        if (selectedText.length > 0) {
            linkText = selectedText;
        } else {
            linkText = prompt('Link Text:', '');
        }
        let newTab = prompt('Open link in new tab? (yes/no)', 'no');
        if (newTab === "yes" || newTab === "y") {
            linkOptions = "target='_blank' rel='noreferrer noopener'";
        }
    }

    let originalAsset;
    const swapAsset = () => {
        originalAsset = field;
        changingAsset = field;
        showMedia = true;
    }
    $: if (changingAsset) {
        if (field === originalAsset) {
            field = changingAsset;
        }
    }

    // If an img path is 404, load the data image instead
    const loadDataImage = imgEl => {
        // Get src from img that was clicked on in visual editor
        let src = imgEl.target.attributes.src.nodeValue;
        // Load all image on the page with that source
        // TODO: Could load images not related to this field specifically
        let allImg = document.querySelectorAll('img[src="' + src + '"]');
        allImg.forEach(i => {
            localMediaList.forEach(asset => {
                // Check if the field path matches a recently uploaded file in memory
                if(asset.file === field) {
                    // Set the source to the data image instead of the path that can't be found
                    i.src = asset.contents; 
                }
            });
        });
    }

    const heading = level => {
        let s = window.getSelection();
        if (s.baseNode.parentNode.tagName === level.toUpperCase()) {
            document.execCommand('formatBlock', false, 'div');
        } else {
            document.execCommand('insertHTML', false, '<' + level + '>' + s + '</' + level + '>');
        }
    }
</script>

{#if field === null}
    <div class="field">{field} is null</div>
{:else if field === undefined}
    <div class="field">{field} is undefined</div>
{:else if schema && schema.hasOwnProperty(parentKeys)}
    {#if schema[parentKeys].type === "checkbox"}
        <Checkbox {schema} {parentKeys} bind:field />
    {/if}
    {#if false}
        {console.log(parentKeys.split('.').reduce((o,i)=> o[i], schema))}
        <div>WYSIWYG</div>
    {/if}
{:else if typeof field === "number"}
    <div class="field">
        <input id="{label}" type="number" bind:value={field} />
    </div>
{:else if typeof field === "string"}
    {#if isDate(field)}
        <div class="field">
            <input type="date" value={makeDate(field)} on:input={date => bindDate(date.target.value)} />
        </div>
    {:else if isAssetPath(field)}
        <div class="field thumbnail-wrapper">
            {#if isImagePath(field)}
                <img src="{field}" alt="click to change thumbnail" class="thumbnail" on:error={imgEl => loadDataImage(imgEl)} />
            {:else if isDocPath(field)}
                <embed src="{field}" class="thumbnail" />
            {/if}
            <button class="swap" on:click|preventDefault={swapAsset}>Change Asset</button>
        </div>
    {:else if field.length < 50}
        <div class="field">
            <input id="{label}" type="text" bind:value={field} />
        </div>
    {:else}
        <div class="field">
            <div class="editor">
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("bold")} title="Bold the selected text">
                    <b>B</b>
                </button>
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("italic")} title="Italicize the selected text">
                    <i>I</i>
                </button>
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("underline")} title="Underline the selected text">
                    <u>U</u>
                </button>
                <div class="spacer"></div>
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("insertUnorderedList")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-list-numbers-MODIFIED" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M11 6h9" />
                        <path d="M11 12h9" />
                        <path d="M12 18h8" />
                        <circle cx="5" r="2" cy="7"></circle>
                        <circle cx="5" r="2" cy="17"></circle>
                    </svg>
                </button>
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("insertOrderedList")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-list-numbers" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M11 6h9" />
                        <path d="M11 12h9" />
                        <path d="M12 18h8" />
                        <path d="M4 16a2 2 0 1 1 4 0c0 .591 -.5 1 -1 1.5l-3 2.5h4" />
                        <path d="M6 10v-6l-2 2" />
                    </svg>
                </button>
                <div class="spacer"></div>
                <button on:click={textarea.focus()} on:click={createLink} on:click|preventDefault={() => document.execCommand("insertHTML", false, "<a href='" + linkURL + "' " + linkOptions + ">" + linkText + "</a>")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-link" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M10 14a3.5 3.5 0 0 0 5 0l4 -4a3.5 3.5 0 0 0 -5 -5l-.5 .5" />
                        <path d="M14 10a3.5 3.5 0 0 0 -5 0l-4 4a3.5 3.5 0 0 0 5 5l.5 -.5" />
                    </svg>
                </button>
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("unlink")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-unlink" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M10 14a3.5 3.5 0 0 0 5 0l4 -4a3.5 3.5 0 0 0 -5 -5l-.5 .5" />
                        <path d="M14 10a3.5 3.5 0 0 0 -5 0l-4 4a3.5 3.5 0 0 0 5 5l.5 -.5" />
                        <line x1="16" y1="21" x2="16" y2="19" />
                        <line x1="19" y1="16" x2="21" y2="16" />
                        <line x1="3" y1="8" x2="5" y2="8" />
                        <line x1="8" y1="3" x2="8" y2="5" />
                    </svg>
                </button>
                <div class="spacer"></div>
                <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h1")} title="Heading level one">
                    h1
                </button>
                <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h2")} title="Heading level two">
                    h2
                </button>
                <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h3")} title="Heading level three">
                    h3
                </button>
                <div class="spacer"></div>
                <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("removeFormat")}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-clear-formatting" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M17 15l4 4m0 -4l-4 4" />
                        <path d="M7 6v-1h11v1" />
                        <line x1="7" y1="19" x2="11" y2="19" />
                        <line x1="13" y1="5" x2="9" y2="19" />
                    </svg>
                </button>
            </div>
            <div id="{label}" class="textarea" contenteditable=true bind:innerHTML={field} bind:this={textarea}></div>
        </div>
    {/if}
{:else if typeof field === "boolean"}
    <div class="field">
        <input id="{label}" type="checkbox" bind:checked={field} /><span>{field}</span>
    </div>
{:else if field.constructor === [].constructor}
    <div class="dragdroplist">
        <div 
            bind:this={ghost}
            id="ghost"
            class={grabbed ? "item haunting" : "item"}
            style={"top: " + (mouseY + offsetY - layerY) + "px"}><p></p></div>
        <div 
            class="list"
            on:mousemove={function(ev) {ev.stopPropagation(); drag(ev.clientY);}}
            on:touchmove={function(ev) {ev.stopPropagation(); drag(ev.touches[0].clientY);}}
            on:mouseup={function(ev) {ev.stopPropagation(); release(ev);}}
            on:touchend={function(ev) {ev.stopPropagation(); release(ev.touches[0]);}}>
        {#each field as value, key (compID = isOpen ? key : JSON.stringify(value))}
            <div 
                id={(grabbed && compID == grabbed.dataset.id) ? "grabbed" : ""}
                data-index={key}
                data-id={compID}
                data-grabY="0"
                class="item-wrapper"
                animate:flip|local={{duration: isOpen ? null : 200}}
            >
            <div class="item">
                <div
                    class="grip"
                    on:mousedown={function(ev) {grab(ev.clientY, this.closest(".item-wrapper"));}}
                    on:touchstart={function(ev) {grab(ev.touches[0].clientY, this.closest(".item-wrapper"));}}
                    on:mouseenter={function(ev) {ev.stopPropagation(); dragEnter(ev, ev.target.closest(".item-wrapper"));}}
                    on:touchmove={function(ev) {ev.stopPropagation(); ev.preventDefault(); touchEnter(ev.touches[0]);}}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-grip-horizontal" width="30" height="30" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <circle cx="5" cy="9" r="1" />
                        <circle cx="5" cy="15" r="1" />
                        <circle cx="12" cy="9" r="1" />
                        <circle cx="12" cy="15" r="1" />
                        <circle cx="19" cy="9" r="1" />
                        <circle cx="19" cy="15" r="1" />
                    </svg>
                </div>
                <div class="buttons">
                    <button 
                        class="up" 
                        style={"visibility: " + (key > 0 ? "" : "hidden") + ";"}
                        on:click|preventDefault={function(ev) {moveItem(key, key - 1)}}>
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16px" height="16px"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6 1.41 1.41z"/></svg>
                    </button>
                    <button 
                        class="down" 
                        style={"visibility: " + (key < field.length - 1 ? "" : "hidden") + ";"}
                        on:click|preventDefault={function(ev) {moveItem(key, key + 1)}}>
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16px" height="16px"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"/></svg>
                    </button>
                </div>

                <div class="content" on:click|preventDefault={accordion(key)}>
                    {#if value.constructor === "".constructor}
                        {value.replace(/<[^>]*>?/gm, '').slice(0, 20).concat(value.length > 20 ? '...' : '')}
                    {:else if value.constructor === ({}).constructor}
                        {Object.values(value)[0].constructor === "".constructor ? Object.values(value)[0].replace(/<[^>]*>?/gm, '').slice(0, 20).concat(value.length > 20 ? '...' : '') : Object.keys(value)[0]}
                    {:else}
                        Component {key}
                    {/if}
                </div>

                <div class="buttons">
                    {#if removesItems}
                        <button
                            class="close"
                            on:click|preventDefault={() => removeItem(value)}>
                            <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 0 24 24" width="16"><path d="M0 0h24v24H0z" fill="none"/><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                        </button>
                    {/if}
                </div>
            </div>
            {#if openKeys.includes(key)}
                <div transition:slide={{ duration: 300 }}>
                    <svelte:self bind:field={field[key]} {label} bind:showMedia bind:changingAsset bind:localMediaList parentKeys={parentKeys + '.' + key} {schema} />
                </div>
            {/if}
            </div>
        {/each}
        </div>
    </div>
{:else if field.constructor === ({}).constructor}
    <fieldset>
        <!-- <legend>{label}</legend> -->
        {#each Object.entries(field) as [key, value]}
            <div class="field">
                <label for={key}>{key}</label>
                <svelte:self bind:field={field[key]} label={key} bind:showMedia bind:changingAsset bind:localMediaList parentKeys={parentKeys + '.' + key} {schema} />
            </div>
        {/each}
    </fieldset>
{/if}

<style>
    label {
        display: block;
    }
    .field {
        margin-bottom: 20px;
    }
    .field:last-of-type {
        margin-bottom: 0;
    }
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
        width: 100%;
        box-sizing: border-box;
    }
    .editor {
        display: flex;
    }
    .editor button {
        background: transparent;
        border: transparent;
        padding: 8px;
        cursor: pointer;
    }
    .editor button:hover {
        background: gray;
    }
    .editor svg {
        display: flex;
        align-content: center;
    }
    .spacer {
        width: 1px;
        background: #777;
        margin: 5px 10px;
    }
    .textarea {
        background: white;
        border: 1px solid gainsboro;
        resize: vertical;
        overflow: auto;
        padding: 7px;
        font-family: sans-serif;
        font-size: small;
    }
    .dragdroplist {
        position: relative;
    }
    .list {
        z-index: 5;
        display: flex;
        flex-direction: column;
    }
    .item-wrapper {
        background-color: gainsboro;
        margin-bottom: 0.5em;
    }
    .item {
        box-sizing: border-box;
        display: inline-flex;
        width: 100%;
        min-height: 2em;
        background-color: white;
        border: 1px solid rgb(190, 190, 190);
        border-radius: 2px;
        user-select: none;
    }
    .item:last-child {
        margin-bottom: 0;
    }
    .item:not(#grabbed):not(#ghost) {
        z-index: 10;
    }
    .item > .buttons {
        margin: auto;
    }
    .content {
        cursor: pointer;
        display: flex;
        flex-grow: 1;
        align-items: center;
        justify-content: center;
    }
    .grip {
        margin-left: 5px;
        display: flex;
        align-items: center;
        cursor: grab;
    }
    .buttons {
        width: 32px;
        min-width: 32px;
        margin: auto 0;
        display: flex;
        flex-direction: column;
    }
    .buttons button {
        cursor: pointer;
        height: 18px;
        margin: 0 auto;
        padding: 0;
        border: 1px solid rgba(0, 0, 0, 0);
        background-color: inherit;
    }
    .buttons button:focus {
        border: 1px solid black;
    }
    #grabbed {
        opacity: 0.0;
    }
    #ghost {
        pointer-events: none;
        z-index: -5;
        position: absolute;
        top: 0;
        left: 0;
        opacity: 0.0;
    }
    #ghost * {
        pointer-events: none;
    }
    #ghost.haunting {
        z-index: 20;
        opacity: 1.0;
    }
    .thumbnail-wrapper {
        height: 115px;
        overflow: hidden;
        position: relative;
    }
    .thumbnail {
        max-width: 200px;
    }
    button.swap {
        cursor: pointer;
        position: absolute;
        top: 0;
        left: 0;
        width: 200px;
        height: 115px;
        border: 0;
        background-color: transparent;
        color: transparent;
        font-size: 1.25rem;
        transition: all .15s;
    }
    button.swap:hover {
        background-color: rgba(0, 0, 0, .75);
        color: white;
    }
</style>