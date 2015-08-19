$(document).ready(function() {
  var pc;
  var isCaller;
  var currentUUID;
  var peerConnections = {};
  var currentStream;
  var ws;

  $.ajax({
    url: "/uuid",
    async: false,
    success: function(data) {
      currentUUID = data;
      var roomName = window.location.pathname.replace("/rooms/", "");
      var url = "ws://" + window.location.host + "/ws?uuid=" + currentUUID + "&room_name=" + roomName;
      ws = new WebSocket(url);

      // get the local stream, show it in the local video element and send it
      navigator.getUserMedia({ "video": true, "audio": true }, gotStream, error);
    }
  });

  ws.onmessage = function (evt) {
    var msg = JSON.parse(evt.data);
    console.log(msg);

    var pc = getPeerConnection(msg.from);

    switch (msg.type) {
      case "peer.connected":
        createOffer(msg.from, pc);
        break;
      case "peer.disconnected":
        $("#" + msg.disconnected).remove();
        break;
      case "icecandidate":
        pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
        break;
      case "sdp.offer":
        pc.setRemoteDescription(new RTCSessionDescription(msg.sdp), function () {
          console.log('Setting remote description by offer');
          pc.createAnswer(function (sdp) {
            pc.setLocalDescription(sdp);
            ws.send(JSON.stringify({
              type: "sdp.answer",
              from: currentUUID,
              to: msg.from,
              sdp: sdp
            }));
          },
          function() { console.log("Answer create successfully") },
          function() { console.log("Creating an answer failed") });
        }, function() { console.log("Setting a remote description failed") });
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

      ws.send(JSON.stringify({ type: "sdp.offer", from: currentUUID, to: id, sdp: sdp }));
    });
  }

  function getPeerConnection(uuid) {
    if (peerConnections[uuid]) {
      return peerConnections[uuid];
    }

    console.log('Creating a new peer connection for', uuid);

    var pc = new RTCPeerConnection();
    peerConnections[uuid] = pc;

    pc.addStream(currentStream);

    pc.onicecandidate = function (evt) {
      if (evt.candidate) {
        ws.send(JSON.stringify({
          type: "icecandidate",
          from: currentUUID,
          to: uuid,
          candidate: evt.candidate
        }));
      }
    };

    pc.onaddstream = function (evt) {
      var video = document.createElement('video');
      video.src = URL.createObjectURL(evt.stream);
      video.className = "video";
      video.autoPlay = true;
      video.play();

      var htmlElem = $("<div id='" + uuid + "' class='col-md-4'></div>");
      htmlElem.append(video);
      $("#remote-videos").append(htmlElem);
      setNickname(uuid);
    };

    return pc;
  }

  function gotStream(stream) {
    $("#local-video")[0].src = URL.createObjectURL(stream);
    currentStream = stream;

    ws.send(JSON.stringify({ type: "peer.connected", from: currentUUID, to: "*" }));
  }

  function error() {
    console.log("error");
  }

  function setNickname(uuid) {
    $.ajax({
      url: "http://www.whimsicalwordimal.com/api/name",
      success: function(data) {
        var span = $("<span>" + data.name + ":</span>");
        $("#" + uuid).prepend(span);
      }
    });
  }
});
