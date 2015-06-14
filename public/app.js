$(document).ready(function() {
  var ws = new WebSocket('ws://localhost:8080/ws');
  window.ws = ws;
  var pc;
  var isCaller;

  ws.onmessage = function (evt) {
    if (!pc) {
      start();
    }

    var signal = JSON.parse(evt.data);

    if (signal.sdp) {
      pc.setRemoteDescription(new RTCSessionDescription(signal.sdp));
    } else {
      pc.addIceCandidate(new RTCIceCandidate(signal.candidate));
    }
  };

  // run start(true) to initiate a call
  function start() {
    pc = new RTCPeerConnection();

    // send any ice candidates to the other peer
    pc.onicecandidate = function (evt) {
      if (evt.candidate) {
        ws.send(JSON.stringify({ "candidate": evt.candidate }));
      }
    };

    // once remote stream arrives, show it in the remote video element
    pc.onaddstream = function (evt) {
      var video = document.createElement('video');
      video.src = URL.createObjectURL(evt.stream);
      video.className = "video";
      video.autoPlay = true;
      video.play();

      $("#remote-videos").append(video);
    };

    // get the local stream, show it in the local video element and send it
    navigator.getUserMedia({ "video": true }, gotStream, error);
  }

  function gotStream(stream) {
    $("#local-video")[0].src = URL.createObjectURL(stream);
    pc.addStream(stream);

    if (isCaller) {
      pc.createOffer(gotDescription);
    } else {
      pc.createAnswer(gotDescription);
    }
  }

  function error() {
    console.log("error");
  }

  function gotDescription(desc) {
    pc.setLocalDescription(desc);
    ws.send(JSON.stringify({ "sdp": desc }));
  }

  $("#start_btn").click(function() {
    isCaller = true;
    start();
  });
});
