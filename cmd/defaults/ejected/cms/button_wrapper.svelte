<div class="buttons">
    <slot></slot>
</div>

<style>
    .buttons {
        display: flex;
        gap: 20px;
        width: 100%;
    }
    .buttons > :global(*) {
        width: 100%;
    }
    .buttons :global(button) {
        width: 100%;
        border-radius: 6px;
        cursor: pointer;
    }
    .buttons :global(button),
    .buttons :global(.button-primary button) {
        background-color: #1c7fc7;
        border: none;
        color: #fff;
        font-weight: bold;
        line-height: 21px;
        padding: 10px;
    }
    .buttons > :global(button:last-child),
    .buttons :global(.button-secondary button) {
        background-color: transparent;
        border: 2px solid #1c7fc7;
        color: #1c7fc7;
    }
</style>