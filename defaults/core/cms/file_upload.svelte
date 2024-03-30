<script>
    import MediaFilters from './media_filters.svelte';
    import MediaGrid from './media_grid.svelte';
    import ButtonWrapper from './button_wrapper.svelte';
    import Button from './button.svelte';

    export let media, changingMedia, showMediaModal, localMediaList, mediaPrefix, user;
    let enabledFilters = [];

    const createMediaList = file => {
        let reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = e => {
            localMediaList = [...localMediaList, {
                file: mediaPrefix + "media/" + file.name,
                contents: e.target.result
            }];
        };
    }
    const selectFile = files => {
        Array.from(files).forEach(file => {
            createMediaList(file);
        });
    }

    let filePrefix = mediaPrefix + "media/";
    $: if (enabledFilters) {
        if (enabledFilters.length > 0) {
            // Convert filter array to path
            let filterPath = enabledFilters[0].join('/') + "/";
            let newPrefix = mediaPrefix + "media/" + filterPath;
            localMediaList.forEach(mediaFile => {
                mediaFile.file = mediaFile.file.replace(filePrefix, newPrefix);
            });
            // Set new prefix in case filter is switched and needs to be replaced
            filePrefix = newPrefix;
        }
    }

    let drag;
    const toggleDrag = () => {
        drag = !drag;
    }
    const dropFile = ev => {
        if (ev.dataTransfer.items) {
            // Use DataTransferItemList interface to access the file(s)
            for (let i = 0; i < ev.dataTransfer.items.length; i++) {
                // If dropped items aren't files, reject them
                if (ev.dataTransfer.items[i].kind === 'file') {
                    let file = ev.dataTransfer.items[i].getAsFile();
                    createMediaList(file);
                }
            }
        }
    }

    let selectedMedia = [];
    const removeSelectedMedia = () => {
        selectedMedia.forEach(file => {
            localMediaList = localMediaList.filter(i => i.contents !== file);
            selectedMedia = [];
        });
    }

    const getThumbnails = mediaList => mediaList.map(i => i.contents);

    const addUploadToLibrary = () => {
        localMediaList.forEach(m => {
            media = [...media, m.contents];
        });
    }
</script>

<div class="upload-wrapper">
    {#if localMediaList.length > 0}
        <MediaFilters bind:media bind:enabledFilters singleSelect={true} {changingMedia} />
        <MediaGrid files={getThumbnails(localMediaList)} bind:selectedMedia={selectedMedia} />
        <ButtonWrapper>
            <Button 
                on:click={() => {
                    addUploadToLibrary();
                    enabledFilters=[];
                    filePrefix = mediaPrefix + "media/";
                    if(changingMedia) {
                        changingMedia = localMediaList[0].file;
                        showMediaModal = false;
                    }
                }}
                bind:commitList={localMediaList}
                buttonText="Save Media"
                action="create"
                encoding="base64"
                {user}
            />
            {#if selectedMedia.length > 0}
                <Button
                    on:click="{removeSelectedMedia}"
                    buttonText="Discard selected"
                    buttonStyle="secondary"
                />
            {:else}
                <Button
                    on:click="{() => localMediaList=[]}"
                    buttonText="Discard all"
                    buttonStyle="secondary"
                />
            {/if}
        </ButtonWrapper>
    {:else}
        <div class="upload-widgets">
            <div class="drop{drag ? ' active' : ''}"
                on:dragenter={toggleDrag} 
                on:dragleave={toggleDrag}  
                on:drop|preventDefault={event => dropFile(event)} 
                on:dragover|preventDefault
            >
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
            <div class="choose" on:change={event => selectFile(event.target.files)}>
                <label class="file">
                    <input type="file" multiple="{changingMedia ? false : true}" aria-label="File browser">
                    <span class="file-custom"></span>
                </label>
            </div>
        </div>
    {/if}
</div>

<style>
    .upload-wrapper {
        display: flex;
        flex-direction: column;
        overflow: hidden;
        height: 100%;
    }
    .upload-widgets {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        box-sizing: border-box;
    }
    .drop {
        width: 100%;
        height: 40%;
        box-sizing: border-box;
        justify-content: center;
        border: 2px dashed;
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .drop.active {
        border-color: #1c7fc7;
        background-color: gainsboro;
    }
    .or {
        margin: 20px;
    }
    .file {
        position: relative;
        cursor: pointer;
    }
    .file input {
        border-radius: 50%;
    }
    .file-custom {
        position: absolute;
        top: 0;
        right: 0;
        left: 0;
        z-index: 5;
        padding: 0.5rem 1rem;
        background-color: #fff;
        border: 0.075rem solid #ddd;
        border-radius: 0.25rem;
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
        content: "Browse";
        padding: 0.5rem 1rem;
        background-color: #eee;
        border: 0.075rem solid #ddd;
        border-radius: 0 0.25rem 0.25rem 0;
    }
    .file-custom:after {
        content: "Choose file...";
    }
</style>