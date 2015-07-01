package main

import (
	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"
)

var rooms map[string]map[string]Peer

type Peer struct {
	// UUID
	uuid string

	// The websocket connection.
	ws *websocket.Conn
}

func addPeer(roomName string, peer Peer) {
	if len(rooms) == 0 {
		rooms = make(map[string]map[string]Peer)
	}

	if len(rooms[roomName]) == 0 {
		peers := make(map[string]Peer)
		rooms[roomName] = peers
	}

	rooms[roomName][peer.uuid] = peer
	go handlePeer(roomName, peer)
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
			for _, oPeer := range rooms[roomName] {
				// Send message to all peers, but not to the sender
				if peer.uuid != oPeer.uuid {
					oPeer.ws.WriteMessage(messageType, p)
				}
			}
		} else {
			rooms[roomName][to].ws.WriteMessage(messageType, p)
		}
	}
}
