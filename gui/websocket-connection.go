package gui

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *Hub
}

func (c *connection) reader() {
	defer c.ws.Close()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		c.h.Broadcast <- message
	}
}

func (c *connection) writer() {
	defer c.ws.Close()

	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type wsHandler struct {
	h *Hub
}

//func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//fmt.Println("New socket")

//ws, err := upgrader.Upgrade(w, r, nil)
//if err != nil {
//return
//}
//c := &connection{send: make(chan []byte, 256), ws: ws, h: wsh.h}
//c.h.register <- c
//defer func() { c.h.unregister <- c }()
//fmt.Println("New socket - registrer")

//go c.writer()
//go c.sendFileList()

//fmt.Println("New socket - reader")
//c.reader()
//}
var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Wshandler(w http.ResponseWriter, r *http.Request, hub *Hub) {
	ws, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, h: hub}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()

	go c.sendCategories(hub.Categories)
	go c.sendFileList()

	c.reader()
}
