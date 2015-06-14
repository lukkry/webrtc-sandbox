$(document).ready(function() {
  var ws = new WebSocket('ws://localhost:8080/ws');
  var pc;
  var isCaller;
  var currentId = generateId();
  var peerConnections = {};
  var currentStream;

  $("#clients").append(currentId);

  // get the local stream, show it in the local video element and send it
  navigator.getUserMedia({ "video": true }, gotStream, error);

  ws.onmessage = function (evt) {
    var msg = JSON.parse(evt.data);
    var pc = getPeerConnection(msg.from);

    console.log(msg);

    switch (msg.type) {
      case "peer.connected":
        createOffer(msg.from, pc);
        $("#clients").append(msg.from);
        break;
      case "icecandidate":
        pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
        break;
      case "sdp.offer":
        pc.setRemoteDescription(new RTCSessionDescription(msg.sdp), function () {
          console.log('Setting remote description by offer');
          pc.createAnswer(function (sdp) {
            pc.setLocalDescription(sdp);
            ws.send(JSON.stringify({ type: "sdp.answer", from: currentId, to: msg.id, sdp: sdp }));
          });
        });
        break;
      case 'sdp.answer':
        pc.setRemoteDescription(new RTCSessionDescription(msg.sdp), function () {
          console.log('Setting remote description by answer');
        }, function (e) {
          console.error(e);
        });
        break;
    }

  };

  function createOffer(id, pc) {
    pc.createOffer(function (sdp) {
      pc.setLocalDescription(sdp);
      console.log('Creating an offer for', id);

      ws.send(JSON.stringify({ type: "sdp.offer", from: currentId, to: id, sdp: sdp }));
    });
  }

  function getPeerConnection(id) {
    if (peerConnections[id]) {
      return peerConnections[id];
    }

    console.log('Creating a new peer connection for', id);

    var pc = new RTCPeerConnection();
    peerConnections[id] = pc;

    pc.addStream(currentStream);

    pc.onicecandidate = function (evt) {
      if (evt.candidate) {
        ws.send(JSON.stringify({ type: "icecandidate", from: currentId, to: id, candidate: evt.candidate }));
      }
    };

    pc.onaddstream = function (evt) {
      var video = document.createElement('video');
      video.src = URL.createObjectURL(evt.stream);
      video.className = "video";
      video.autoPlay = true;
      video.play();

      $("#remote-videos").append(video);
    };

    return pc;
  }

  function gotStream(stream) {
    $("#local-video")[0].src = URL.createObjectURL(stream);
    currentStream = stream;

    ws.send(JSON.stringify({ type: "peer.connected", from: currentId }));
  }

  function error() {
    console.log("error");
  }

  function generateId() {
    return '_' + Math.random().toString(36).substr(2, 9);
  }
});
