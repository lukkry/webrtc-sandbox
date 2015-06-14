package main

import (
	"strings"

	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var dialer = websocket.Dialer{}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func createConnection(t *testing.T, url string) *websocket.Conn {
	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	return ws
}

func sendMsg(t *testing.T, ws *websocket.Conn, msg string) {
	if err := ws.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetWriteDeadline: %v", err)
	}
	if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
}

func assertMsgReceived(t *testing.T, ws *websocket.Conn, msg string) {
	if err := ws.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(p) != msg {
		t.Fatalf("message=%s, want %s", p, msg)
	}
}

func assertMsgNotReceived(t *testing.T, ws *websocket.Conn) {
	if err := ws.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	_, p, err := ws.ReadMessage()
	if err == nil {
		t.Fatalf("ReadMessage: %v", err)
		t.Fatalf("message should not be received %s", p)
	}
}

func TestMessageIsSendToAllPeers(t *testing.T) {
	ts := httptest.NewServer(&WsHandler{})
	ts.URL = makeWsProto(ts.URL)
	setup()
	defer ts.Close()

	conn1 := createConnection(t, ts.URL)
	defer conn1.Close()

	conn2 := createConnection(t, ts.URL)
	defer conn2.Close()

	conn3 := createConnection(t, ts.URL)
	defer conn3.Close()

	const message = "Hello World!"

	sendMsg(t, conn1, message)
	assertMsgNotReceived(t, conn1)
	assertMsgReceived(t, conn2, message)
	assertMsgReceived(t, conn3, message)
}
