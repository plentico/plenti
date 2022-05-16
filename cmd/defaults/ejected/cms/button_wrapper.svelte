<div class="buttons">
    <slot></slot>
</div>

<style>
    .buttons {
        display: flex;
        gap: 20px;
        width: 100%;
    }
    .buttons :global(button) {
        background-color: #1c7fc7;
        border: none;
        border-radius: 6px;
        color: #fff;
        cursor: pointer;
        font-weight: bold;
        line-height: 21px;
        padding: 10px;
        width: 100%;
    }
    .buttons :global(button:last-of-type) {
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
        width: 100%;
        border-radius: 6px;
        cursor: pointer;
    }
</style>