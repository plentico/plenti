<script>
    import { isDate, makeDate, formatDate } from './dates.js';
    export let field, label;

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
</script>

{#if field === null}
    <div>{field} is null</div>
{:else if field === undefined}
    <div>{field} is undefined</div>
{:else if field.constructor === "".constructor}
    {#if isDate(field)}
        <input type="date" value={makeDate(field)} on:input={date => bindDate(date.target.value)} />
    {:else if field.length < 50}
        <input id="{label}" type="text" bind:value={field} />
    {:else}
        <textarea id="{label}" rows="5" bind:value={field}></textarea>
    {/if}
{:else if field.constructor === true.constructor}
    <input id="{label}" type="checkbox" bind:checked={field} /><span>{field}</span>
{:else if field.constructor === [].constructor}
    <main class="dragdroplist">
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
        {#each field as value, key (compID = isOpen ? key : value.constructor === ({}).constructor ? value[Object.keys(value)[0]] : value)}
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
                            on:click|preventDefault={() => removeItem(value)}>
                            <svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 0 24 24" width="16"><path d="M0 0h24v24H0z" fill="none"/><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                        </button>
                    {/if}
                </div>
            </div>
            {#if openKeys.includes(key)}
                <div transition:slide={{ duration: 300 }}>
                    <svelte:self bind:field={field[key]} {label} />
                </div>
            {/if}
            </div>
        {/each}
        </div>
    </main>
{:else if field.constructor === ({}).constructor}
    <fieldset>
        <!-- <legend>{label}</legend> -->
        {#each Object.entries(field) as [key, value]}
            <div>
                <label for={key}>{key}</label>
                <svelte:self bind:field={field[key]} label={key} />
            </div>
        {/each}
    </fieldset>
{/if}

<style>
    label {
        display: block;
    }
    input[type=text] {
        height: 30px;
        padding: 0 7px;
        border: 1px solid gainsboro;
        border-radius: 3px;
        width: 100%;
        box-sizing: border-box;
    }
    textarea {
        box-sizing: border-box;
        width: 100%;
    }
    main {
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
</style>