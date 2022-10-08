<script>
    import DynamicFormInput from "../dynamic_form_input.svelte";
    import allComponentDefaults from "../../component_defaults.js";
    import allComponentSchemas from '../../component_schemas.js';
    export let field, label, showMedia, changingAsset, localMediaList, parentKeys, schema;

    const objKeysMatch = (a, b) => {
        var aKeys = Object.keys(a).sort();
        var bKeys = Object.keys(b).sort();
        return JSON.stringify(aKeys) === JSON.stringify(bKeys);
    }
    let compSchema;
    const setCompSchema = component => {
        let compDefaults = structuredClone(allComponentDefaults);
        let compSchemas = structuredClone(allComponentSchemas);
        // Deep clone so we're not changing original component
        let b = structuredClone(component);
        // Temp remove salt for comparison
        delete b.plenti_salt;
        for (const c in compDefaults) {
            if (objKeysMatch(compDefaults[c], b)) {
                compSchema = compSchemas[c];
            }
        }
    }

    // Accordion
    import {slide} from "svelte/transition";
    let isOpen = false;
    let openKeys = [];
    const accordion = (newKey, component) => {
        setCompSchema(component);
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
    let compList = false;
    const toggleCompList = () => {
        compList = !compList;
    }
    let addName;
    const addComponent = component => {
        let compDefaults = structuredClone(allComponentDefaults);
        // Check if there is a component default available
        if (component in compDefaults) {
            field.forEach(c => {
                // Check if exact component value exists on page already
                if (JSON.stringify(c) === JSON.stringify(compDefaults[component])) {
                    compDefaults[component].plenti_salt = createSalt();
                }
            });
            field = [...field, compDefaults[component]];
            addName = component;
        } else {
            addName = component + "not_found";
        }
        setTimeout(() => {
            addName = "";
        }, 250);
    }
    const createSalt = () => {
        // Create salt give duplicate components some uniqueness
        return (Math.random() + 1).toString(36).substring(7);
    }
    const toggleSalt = component => {
        if ('plenti_salt' in component) {
            for (const c of field) {
                // Deep clone so we're not changing original component
                let b = structuredClone(component);
                // Temp remove salt for comparison
                delete b.plenti_salt;
                // Check if exact component value exists on page already
                if (JSON.stringify(c) === JSON.stringify(b)) {
                    // Still matching, keep salt
                    return;
                }
            };
            // No matches, remove salt
            delete component.plenti_salt;
            component = component;
        } else {
            // Check if component matches more than itself
            if (field.filter(c => JSON.stringify(c) === JSON.stringify(component)).length > 1) {
                component.plenti_salt = createSalt();
            }
        }
    }
</script>

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

            <div 
                class="content" 
                on:click|preventDefault={accordion(key, value)}
                on:click={toggleSalt(value)}
            >
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
            <div transition:slide|local={{ duration: 300 }}>
                <DynamicFormInput
                    bind:field={field[key]}
                    label={null}
                    bind:showMedia
                    bind:changingAsset
                    bind:localMediaList
                    parentKeys={""}
                    {schema}
                    {compSchema}
                />
            </div>
        {/if}
        </div>
    {/each}
    </div>
    {#if schema && schema[parentKeys]?.options.length > 0}
        <button class="add{compList ? ' open':''}" on:click|preventDefault={toggleCompList}>
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-circle-plus" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="#1c7fc7" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                <circle cx="12" cy="12" r="9"></circle>
                <line x1="9" y1="12" x2="15" y2="12"></line>
                {#if !compList}
                    <line x1="12" y1="9" x2="12" y2="15"></line>
                {/if}
            </svg>
            Add new {label}
        </button>
        {#if compList}
            <div class="add-list" transition:slide|local={{ duration: 300 }}>
                {#each schema[parentKeys].options as option}
                    <button 
                        class="add-name"
                        on:click|preventDefault={() => addComponent(option)}
                    >
                        {#if addName === option}
                            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-check" width="15" height="15" viewBox="0 0 24 24" stroke-width="2" stroke="#4bb543" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                <path d="M5 12l5 5l10 -10"></path>
                            </svg>
                        {:else if addName === option + "not_found"}
                            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="15" height="15" viewBox="0 0 24 24" stroke-width="2" stroke="#ed0f0f" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                <line x1="18" y1="6" x2="6" y2="18"></line>
                                <line x1="6" y1="6" x2="18" y2="18"></line>
                            </svg>
                        {/if}
                        {option}
                    </button>
                {/each}
            </div>
        {/if}
    {/if}
</div>

<style>
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
        border: 1px solid gainsboro;
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
    .add, .add-name {
        background-color: transparent;
        border: none;
        cursor: pointer;
        display: block;
    }
    .add {
        display: flex;
        gap: 5px;
        padding-left: 4px;
        align-items: center;
        background-color: white;
        border: 1px solid gainsboro;
        position: relative;
        z-index: 1;
    }
    .add.open {
        border-bottom: none;
    }
    .add-name {
        border-radius: 5px;
        padding: 5px;
        border: 1px solid gainsboro;
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 5px;
    }
    .add-list {
        background-color: white;
        border: 1px solid gainsboro;
        margin-top: -1px;
        display: grid;
        grid-template-columns: 1fr 1fr;
        padding: 10px;
        gap: 5px;
    }
</style>