let socket;

let pop = document.createElement('div');
pop.style.cssText = 'font-size: 18px;position:absolute;top:60px;right:10px;padding:30px;color:#fff;z-index:100;border-radius: 8px;background:#599118';

window.onload = () => {
  console.debug("onload");
  
  pop.innerHTML= "Reload Connected..."
  document.body.appendChild(pop);
  setTimeout(function(){
	  pop.style.display="none";
  },1500)
  

};
document.addEventListener("DOMContentLoaded", () => {
  socket = new WebSocket('ws://' + window.location.host + '/reload');

  socket.onmessage = function (event) {
    var datastr = event.data.split(":")[0];
    if (datastr === 'reload'){
      console.debug("reloading");
      window.location.reload();
    }
  };
    
  socket.onopen = function() {
    console.debug("connected");
    socket.send("loaded");
  }
  
  socket.onclose = function(e) {
    pop.innerHTML= "Reload Disconnected...";
    pop.style.display="block";
    pop.style.backgroundColor="rgb(236,64,3)";
    console.debug("connection closed (" + e.code + ")");
  }
  
});