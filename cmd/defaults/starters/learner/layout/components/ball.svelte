<!-- From https://codepen.io/surjithctly/pen/yedHk -->
<div class="ball-wrapper">
  <div class="bouncingball"></div>
</div>

<style>
  .ball-wrapper {
    position: relative;
    height: 80px;
  }
  .bouncingball {
    width: 40px;
    height: 40px;
    border-radius: 100%;
    background: var(--primary);
    animation: bounce 1s;
    transform: translateY(0px);
    animation-iteration-count: infinite;
    position: absolute;
  }

  @keyframes bounce {
    0% {
      top: 0;
      -webkit-animation-timing-function: ease-in;
    }
    40% {}
    50% {
      top: 40px;
      height: 40px;
      -webkit-animation-timing-function: ease-out;
    }
    55% {
      top: 60px; height: 20px; 
      -webkit-animation-timing-function: ease-in;}
    65% {
      top: 20px; height: 40px; 
      -webkit-animation-timing-function: ease-out;}
    95% {
      top: 0;		
      -webkit-animation-timing-function: ease-in;
    }
    100% {
      top: 0;
      -webkit-animation-timing-function: ease-in;
    }
  }

</style>