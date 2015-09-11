### WebRTC Sandbox
Group conversations over WebRTC. See it in action on [webrtc.lukkry.info](http://www.webrtc.lukkry.info). It was tested only with Chrome, but should work on any browser which supports WebRTC.

### How to run it?
1. Run a binary

```shell
$ godep go run webrtc.go hub.go
```

2. Create a tunnel to localhost. Personally, I use [ngrok](http://ngrok.io), but it could be anything.
```shell
$ ngrok http 8080
```

### TODO
* Send files over WebRTC using RTCDataChannel
