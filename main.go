package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WsHandler struct{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var connections map[*websocket.Conn]bool

func (h *WsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connections[conn] = true
	go handleConn(conn)
}

func handleConn(conn *websocket.Conn) error {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		for c := range connections {
			// Send message to all peers, but not to the sender
			if c != conn {
				c.WriteMessage(messageType, p)
			}
		}
	}
}

func setup() {
	connections = make(map[*websocket.Conn]bool)
}

func main() {
	setup()

	http.Handle("/ws", &WsHandler{})
	http.Handle("/", http.FileServer(http.Dir("public")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}