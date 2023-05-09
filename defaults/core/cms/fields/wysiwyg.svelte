<script>
    export let schema, parentKeys, field;

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

    const heading = level => {
        let s = window.getSelection();
        if (s.baseNode.parentNode.tagName === level.toUpperCase()) {
            document.execCommand('formatBlock', false, 'div');
        } else {
            document.execCommand('insertHTML', false, '<' + level + '>' + s + '</' + level + '>');
        }
    }

    const options = option => schema[parentKeys].options.includes(option);
</script>

<div class="editor">
    {#if options("bold") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("bold")} title="Bold the selected text">
            <b>B</b>
        </button>
    {/if}
    {#if options("italic") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("italic")} title="Italicize the selected text">
            <i>I</i>
        </button>
    {/if}
    {#if options("underline") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("underline")} title="Underline the selected text">
            <u>U</u>
        </button>
    {/if}
    {#if options("bold") || options("italic") || options("underline") || options("all")}
        <div class="spacer"></div>
    {/if}
    {#if options("bullets") || options("all")}
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
    {/if}
    {#if options("numbers") || options("all")}
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
    {/if}
    {#if options("bullets") || options("numbers") || options("all")}
        <div class="spacer"></div>
    {/if}
    {#if options("link") || options("all")}
        <button on:click={textarea.focus()} on:click={createLink} on:click|preventDefault={() => document.execCommand("insertHTML", false, "<a href='" + linkURL + "' " + linkOptions + ">" + linkText + "</a>")}>
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-link" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <path d="M10 14a3.5 3.5 0 0 0 5 0l4 -4a3.5 3.5 0 0 0 -5 -5l-.5 .5" />
                <path d="M14 10a3.5 3.5 0 0 0 -5 0l-4 4a3.5 3.5 0 0 0 5 5l.5 -.5" />
            </svg>
        </button>
    {/if}
    {#if options("unlink") || options("all")}
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
    {/if}
    {#if options("link") || options("unlink") || options("all")}
        <div class="spacer"></div>
    {/if}
    {#if options("heading1") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h1")} title="Heading level one">
            h1
        </button>
    {/if}
    {#if options("heading2") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h2")} title="Heading level two">
            h2
        </button>
    {/if}
    {#if options("heading3") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h3")} title="Heading level three">
            h3
        </button>
    {/if}
    {#if options("heading4") || options("all")}
        <button on:click={textarea.focus()} on:click|preventDefault={() => heading("h4")} title="Heading level four">
            h4
        </button>
    {/if}
    {#if options("heading1") || options("heading2") || options("heading3") || options("heading4") || options("all")}
        <div class="spacer"></div>
    {/if}
    {#if options("clear") || options("all")}
    <button on:click={textarea.focus()} on:click|preventDefault={() => document.execCommand("removeFormat")}>
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-clear-formatting" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="black" fill="none" stroke-linecap="round" stroke-linejoin="round">
            <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
            <path d="M17 15l4 4m0 -4l-4 4" />
            <path d="M7 6v-1h11v1" />
            <line x1="7" y1="19" x2="11" y2="19" />
            <line x1="13" y1="5" x2="9" y2="19" />
        </svg>
    </button>
    {/if}
</div>

<div class="textarea" contenteditable=true bind:innerHTML={field} bind:this={textarea}></div>

<style>
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
</style>