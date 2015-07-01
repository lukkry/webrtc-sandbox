package main

import (
	"encoding/json"
	"math/rand"
	"strings"

	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var dialer = websocket.Dialer{}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http") + "/ws"
}

func createConnection(t *testing.T, url string, uuid string) *websocket.Conn {
	url = url + "?uuid=" + uuid
	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	return ws
}

func sendMsg(t *testing.T, ws *websocket.Conn, msg []byte) {
	if err := ws.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetWriteDeadline: %v", err)
	}
	if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
}

func assertMsgReceived(t *testing.T, ws *websocket.Conn, msg []byte) {
	if err := ws.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(p) != string(msg) {
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
	defer ts.Close()

	uuid1 := string(rand.Intn(10000))
	conn1 := createConnection(t, ts.URL, uuid1)
	defer conn1.Close()

	uuid2 := string(rand.Intn(10000))
	conn2 := createConnection(t, ts.URL, uuid2)
	defer conn2.Close()

	uuid3 := string(rand.Intn(10000))
	conn3 := createConnection(t, ts.URL, uuid3)
	defer conn3.Close()

	broadcastMsg, _ := json.Marshal(map[string]string{"type": "peer.connected", "to": "*"})
	sendMsg(t, conn1, broadcastMsg)
	assertMsgNotReceived(t, conn1)
	assertMsgReceived(t, conn2, broadcastMsg)
	assertMsgReceived(t, conn3, broadcastMsg)

	directMsg, _ := json.Marshal(map[string]string{"type": "peer.connected", "to": uuid3})
	sendMsg(t, conn1, directMsg)
	assertMsgNotReceived(t, conn1)
	assertMsgNotReceived(t, conn2)
	assertMsgReceived(t, conn3, directMsg)
}
