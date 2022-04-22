<label class="file">
  <input type="file" id="file" aria-label="File browser">
  <span class="file-custom"></span>
</label>

<style>
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
        height: 2.5rem;
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
        height: 2.5rem;
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
</style>