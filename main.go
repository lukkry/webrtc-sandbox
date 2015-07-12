package main

import (
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

func (h *WsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	uuid := req.URL.Query().Get("uuid")
	roomName := req.URL.Query().Get("room_name")

	peer := Peer{uuid: uuid, roomName: roomName, ws: conn}
	hub.register <- peer
}

func rooms(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "views/room.html")
}

func uuid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, generateUUID())
}

func generateUUID() string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(string(out))
}

func main() {
	RunHub()

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/ws", &WsHandler{})
	http.HandleFunc("/rooms/", rooms)
	http.HandleFunc("/uuid", uuid)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}