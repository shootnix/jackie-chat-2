package daemon

import (
	"github.com/gin-gonic/gin"
	"github.com/shootnix/jackie-chat-2/config"
	"github.com/shootnix/jackie-chat-2/controllers"
)

type Daemon struct {
	r      *gin.Engine
	listen string
}

func NewDaemon() *Daemon {
	r := gin.Default()
	auth := r.Group("/")
	auth.Use(controllers.AuthRequired())
	{
		auth.POST("/api/v1/sendMessage", controllers.SendMessage)
		auth.GET("/api/v1/statusMessage", controllers.GetMessageStatus)
	}

	r.GET("/", controllers.Index)
	r.GET("/ping", controllers.Ping)
	r.POST("/api/v1/auth", controllers.Auth)

	d := &Daemon{r, config.GetConfig().Daemon.Listen}

	return d
}

func (d *Daemon) Run() {
	d.r.Run(d.listen)
}
