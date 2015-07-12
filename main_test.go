package main

import (
	"encoding/json"
	"math/rand"
	"strconv"
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

func createConnection(t *testing.T, url string) *websocket.Conn {
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
	if strings.TrimSpace(string(p)) != string(msg) {
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

func generatePeer(t *testing.T, roomName string, ts *httptest.Server) Peer {
	uuid := strconv.Itoa(rand.Intn(10000))
	url := ts.URL + "?room_name=" + roomName + "&uuid=" + uuid
	conn := createConnection(t, url)
	return Peer{uuid: uuid, ws: conn}
}

func TestMessageIsSendToAllPeersInARoom(t *testing.T) {
	RunHub()

	ts := httptest.NewServer(&WsHandler{})
	ts.URL = makeWsProto(ts.URL)
	defer ts.Close()

	peer1 := generatePeer(t, "RoomA", ts)
	peer2 := generatePeer(t, "RoomA", ts)
	peer3 := generatePeer(t, "RoomA", ts)
	peer4 := generatePeer(t, "RoomB", ts)

	broadcastMsg, _ := json.Marshal(map[string]string{"type": "peer.connected", "to": "*"})
	sendMsg(t, peer1.ws, broadcastMsg)
	assertMsgNotReceived(t, peer1.ws)
	assertMsgReceived(t, peer2.ws, broadcastMsg)
	assertMsgReceived(t, peer3.ws, broadcastMsg)
	assertMsgNotReceived(t, peer4.ws)

	directMsg, _ := json.Marshal(map[string]string{"type": "peer.connected", "to": peer3.uuid})
	sendMsg(t, peer1.ws, directMsg)
	assertMsgNotReceived(t, peer1.ws)
	assertMsgNotReceived(t, peer2.ws)
	assertMsgReceived(t, peer3.ws, directMsg)
	assertMsgNotReceived(t, peer4.ws)
}

func TestPeerIsRemoved(t *testing.T) {
	RunHub()

	ts := httptest.NewServer(&WsHandler{})
	ts.URL = makeWsProto(ts.URL)
	defer ts.Close()

	peer1 := generatePeer(t, "RoomA", ts)
	peer2 := generatePeer(t, "RoomA", ts)
	peer3 := generatePeer(t, "RoomA", ts)

	if err := peer1.ws.WriteMessage(websocket.CloseMessage, []byte("")); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	directMsg, _ := json.Marshal(map[string]string{"type": "peer.disconnected",
		"from": "server", "to": "*", "disconnected": peer1.uuid})

	assertMsgNotReceived(t, peer1.ws)
	assertMsgReceived(t, peer2.ws, directMsg)
	assertMsgReceived(t, peer3.ws, directMsg)
}
