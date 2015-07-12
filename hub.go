package main

import (
	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"
)

type Hub struct {
	// Registered peers
	rooms map[string]map[string]Peer

	// Register a new peer
	register chan Peer

	// Unregister existing peer
	unregister chan Peer
}

type Peer struct {
	uuid     string
	roomName string

	// The websocket connection.
	ws *websocket.Conn
}

var hub = Hub{
	rooms:      make(map[string]map[string]Peer),
	register:   make(chan Peer),
	unregister: make(chan Peer),
}

func RunHub() {
	go func() {
		for {
			select {
			case peer := <-hub.register:
				registerPeer(peer)
			case peer := <-hub.unregister:
				unregisterPeer(peer)
			}
		}
	}()
}

func registerPeer(peer Peer) {
	if len(hub.rooms[peer.roomName]) == 0 {
		peers := make(map[string]Peer)
		hub.rooms[peer.roomName] = peers
	}

	hub.rooms[peer.roomName][peer.uuid] = peer
	go handlePeer(peer.roomName, peer)
}

func unregisterPeer(peer Peer) {
	delete(hub.rooms[peer.roomName], peer.uuid)

	for _, oPeer := range hub.rooms[peer.roomName] {
		payload := map[string]string{"type": "peer.disconnected",
			"from": "server", "to": "*", "disconnected": peer.uuid}
		oPeer.ws.WriteJSON(payload)
	}
}

func handlePeer(roomName string, peer Peer) {
	for {
		messageType, p, err := peer.ws.ReadMessage()
		if err != nil {
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(p, &msg); err != nil {
			panic(err)
		}

		to := strings.TrimSpace(msg["to"].(string))

		if to == "*" {
			for _, oPeer := range hub.rooms[roomName] {
				// Send message to all peers, but not to the sender
				if peer.uuid != oPeer.uuid {
					oPeer.ws.WriteMessage(messageType, p)
				}
			}
		} else {
			hub.rooms[roomName][to].ws.WriteMessage(messageType, p)
		}
	}

	hub.unregister <- peer
	peer.ws.Close()
}
