package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
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

func index(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "views/index.html")
}

func getUUID(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, generateUUID())
}

func generateUUID() string {
	return uuid.New()
}

func main() {
	RunHub()

	port := flag.String("port", "8080", "HTTP Port to listen on")
	flag.Parse()

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.Handle("/ws", &WsHandler{})
	http.HandleFunc("/uuid", getUUID)
	http.HandleFunc("/rooms/", rooms)
	http.HandleFunc("/", index)

	log.Println("Starting Server on", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}