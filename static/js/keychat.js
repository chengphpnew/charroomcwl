
function is_weixn(){
    var ua = navigator.userAgent.toLowerCase();
    var iswexin = ua.match(/MicroMessenger/i)=="micromessenger";
    console.log("agent:",ua);
    console.log("iswexin:",iswexin);
    if(iswexin) {
       window.location.href='/weixin';
       return; 
    }else{
        
    }
}

is_weixn()