package gui

import (
	"github.com/stamp/go-dsmobile/types"

	log "github.com/Sirupsen/logrus"
)

type Hub struct {
	Categories *types.Categories `inject:""`

	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	Broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
	}
}

func (h *Hub) Start() error {
	go func() {
		for {
			select {
			case c := <-h.register:
				h.connections[c] = true
				log.WithFields(log.Fields{
					"addr": c.ws.RemoteAddr().String(),
				}).Debug("WS connection registered")
			case c := <-h.unregister:
				log.WithFields(log.Fields{
					"addr": c.ws.RemoteAddr().String(),
				}).Debug("WS connection unregistered")
				if _, ok := h.connections[c]; ok {
					delete(h.connections, c)
					close(c.send)
				}
			case m := <-h.Broadcast:
				for c := range h.connections {
					select {
					case c.send <- m:
					default:
						delete(h.connections, c)
						close(c.send)
					}
				}
			}
		}
	}()

	return nil
}
