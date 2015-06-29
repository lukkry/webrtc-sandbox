package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
)

type WsHandler struct{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var connections map[string]*websocket.Conn

func (h *WsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	uuid := req.URL.Query().Get("uuid")

	connections[uuid] = conn
	go handleConn(conn, uuid)
}

func uuid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, generateUUID())
}

func handleConn(conn *websocket.Conn, uuid string) error {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(p, &msg); err != nil {
			panic(err)
		}

		to := strings.TrimSpace(msg["to"].(string))

		if to == "*" {
			for peerUUID, peerConn := range connections {
				// Send message to all peers, but not to the sender
				if uuid != peerUUID {
					peerConn.WriteMessage(messageType, p)
				}
			}
		} else {
			connections[to].WriteMessage(messageType, p)
		}
	}
}

func generateUUID() string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(string(out))
}

func setup() {
	connections = make(map[string]*websocket.Conn)
}

func main() {
	setup()

	http.Handle("/ws", &WsHandler{})
	http.HandleFunc("/uuid", uuid)
	http.Handle("/", http.FileServer(http.Dir("public")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}