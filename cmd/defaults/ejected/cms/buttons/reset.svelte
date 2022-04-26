 <script>
    import originalContent from '../../content.js';
    const resetContent = () => {
        // TODO: This is broken
        content = originalContent.find(current => current.filename == content.filename)
    }
 </script>
 <button
    class="reset"
    on:click|preventDefault={resetContent}
>
    Reset
</button>
<style>
    .reset {
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        width: 100%;
        border-radius: 6px;
        cursor: pointer;
    }
</style>