$(document).ready(function() {
  $("#rooms").submit(function(e) {
    roomName = $("#room-name").val();
    window.location.href = window.location.href + "/rooms/" + roomName;

    e.preventDefault();
  });
});
