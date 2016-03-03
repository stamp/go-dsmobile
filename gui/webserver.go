package gui

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type Webserver struct {
	Hub  *Hub `inject:""`
	Port string
}

func NewWebserver(port string) *Webserver {
	return &Webserver{Port: port}
}

func (self *Webserver) Start() error {
	go func() {
		gin.SetMode(gin.TestMode)

		r := gin.Default()
		r.LoadHTMLFiles("gui/public/index.html")
		r.StaticFS("/storage", http.Dir("storage"))

		r.GET("/", func(c *gin.Context) {
			c.HTML(200, "index.html", nil)
		})

		r.GET("/socket", func(c *gin.Context) {
			Wshandler(c.Writer, c.Request, self.Hub)
		})

		r.Use(static.Serve("/", static.LocalFile("gui/public", false)))
		log.Infof("Webserver started at :%s", self.Port)
		r.Run(":" + self.Port)
	}()

	return nil
}
