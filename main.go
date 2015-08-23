package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/koding/multiconfig"
)

type Config struct {
	WebPort  string `default:"3001"`
	Ip       string `default:"10.10.100.1"`
	Username string
	Password string
}

func main() {
	// Load the main config
	config := &Config{}
	m := multiconfig.NewWithPath("config.json")
	m.Load(config)
	saveConfigToFile(config)

	ds := NewDsMobile(config.Ip, config.Username, config.Password)

	// Start the ping monitor
	go pinger(ds)

	// Start the hub for websocket messages
	hub := NewHub()
	go hub.Run()

	go func() {
		for {
			select {
			case uuid := <-ds.NewFile:
				type newFileEvent struct {
					Type string
					Uuid string
					Path string
				}

				e := newFileEvent{
					Type: "newFile",
					Uuid: uuid.String(),
					Path: "/storage/" + uuid.String() + ".JPG",
				}

				msg, _ := json.Marshal(e)
				hub.broadcast <- msg
			}

		}
	}()

	// Start up the webserver
	r := gin.Default()
	r.LoadHTMLFiles("public/index.html")
	r.StaticFS("/storage", http.Dir("storage"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/socket", func(c *gin.Context) {
		wshandler(c.Writer, c.Request, hub)
	})

	r.Run(":" + config.WebPort)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request, hub *Hub) {
	ws, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, h: hub}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()
	c.reader()
}

func saveConfigToFile(config *Config) error {
	configFile, err := os.Create("config.json")
	if err != nil {
		return err
	}

	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)

	return nil
}
