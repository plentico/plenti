<script>
    let thumbnail;
    const selectFile =(e)=>{
        let image = e.target.files[0];
        let reader = new FileReader();
        reader.readAsDataURL(image);
        reader.onload = e => {
            thumbnail = e.target.result
        };
    }
</script>

<div class="upload-wrapper">
    {#if thumbnail}
        <img src="{thumbnail}" />
    {:else}
        <div class="drop">
            <div class="drop-icon">
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-cloud-upload" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="#2c3e50" fill="none" stroke-linecap="round" stroke-linejoin="round">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                    <path d="M7 18a4.6 4.4 0 0 1 0 -9a5 4.5 0 0 1 11 2h1a3.5 3.5 0 0 1 0 7h-1" />
                    <polyline points="9 15 12 12 15 15" />
                    <line x1="12" y1="12" x2="12" y2="21" />
                </svg>
            </div>
            <div class="drop-text">Drag a file here to upload</div>
        </div>
        <div class="or">Or</div>
        <div on:click={(e)=>selectFile(e)} on:change={(e)=>selectFile(e)}>
            <label class="file">
            <input type="file" id="file" aria-label="File browser">
            <span class="file-custom"></span>
            </label>
        </div>
    {/if}
</div>

<style>
    .upload-wrapper {
        padding: 20px;
        display: flex;
        flex-direction: column;
        align-items: center;
        height: 100%;
    }
    .drop {
        width: 100%;
        height: 40%;
        justify-content: center;
        border: 2px dashed;
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .or {
        margin: 20px;
    }
    .file {
        position: relative;
        display: inline-block;
        cursor: pointer;
        height: 2.5rem;
    }
    .file input {
        min-width: 14rem;
        margin: 0;
        opacity: 0;
    }
    .file-custom {
        position: absolute;
        top: 0;
        right: 0;
        left: 0;
        z-index: 5;
        padding: 0.5rem 1rem;
        line-height: 1.5;
        color: #555;
        background-color: #fff;
        border: 0.075rem solid #ddd;
        border-radius: 0.25rem;
        box-shadow: inset 0 0.2rem 0.4rem rgb(0 0 0 / 5%);
        -webkit-user-select: none;
        -moz-user-select: none;
        -ms-user-select: none;
        user-select: none;
    }
    .file-custom:before {
        position: absolute;
        top: -0.075rem;
        right: -0.075rem;
        bottom: -0.075rem;
        z-index: 6;
        display: block;
        content: "Browse";
        padding: 0.5rem 1rem;
        line-height: 1.5;
        color: #555;
        background-color: #eee;
        border: 0.075rem solid #ddd;
        border-radius: 0 0.25rem 0.25rem 0;
    }
    .file-custom:after {
        content: "Choose file...";
    }
    img {
        max-width: 200px;
    }
</style>