<!-- From https://codepen.io/teerapuch/pen/vLJXeR -->
<div class="block"></div>

<style>
  .block {
    width: 40px;
    height: 40px;
    background-color: var(--accent);
    animation-name: spin;
    animation-duration: 4000ms;
    animation-iteration-count: infinite;
    animation-timing-function: linear;
  }

  @keyframes spin {
    from {
        transform:rotate(-50deg);
    }
    to {
        transform:rotate(310deg);
    }
  }
</style>