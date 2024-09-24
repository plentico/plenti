<div class="plenti-modal-wrapper" on:click>
    <div class="plenti-modal-close">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-x" width="35" height="35" viewBox="5 5 14 14" stroke-width="1.5" stroke="white" fill="none" stroke-linecap="round" stroke-linejoin="round">
            <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
    </div>
    <slot></slot>
</div>

<style>
    .plenti-modal-wrapper {
      z-index: 99999;
      position: fixed;
      inset: 0px;
      display: flex;
      -webkit-box-pack: center;
      justify-content: center;
      -webkit-box-align: center;
      align-items: center;
      transition: background-color 0.2s ease 0s, opacity 0.2s ease 0s;
      background-color: rgba(0, 0, 0, 0.6);
    }
    .plenti-modal-wrapper :global(.plenti-modal) {
      box-shadow: rgb(68 74 87 / 15%) 0px 4px 12px 0px, rgb(68 74 87 / 25%) 0px 1px 3px 0px;
      background-color: rgb(255, 255, 255);
      border-radius: 5px;
      height: 80%;
      width: 80%;
      max-width: 1200px;
      padding: 20px;
      display: flex;
      overflow: hidden;
    }
    .plenti-modal-close {
      cursor: pointer;
      position: absolute;
      top: 0;
      right: 0;
    }
</style>