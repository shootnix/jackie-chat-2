package daemon

import (
	"github.com/gin-gonic/gin"
	"github.com/shootnix/jackie-chat-2/config"
	"github.com/shootnix/jackie-chat-2/controllers"
	//"io"
	//"log"
	//"os"
)

type Daemon struct {
	r      *gin.Engine
	listen string
}

func NewDaemon() *Daemon {

	//l := log.New(out, prefix, flag)

	/*
		slash := string(os.PathSeparator)
		debugLogFilePath := "log" + slash + "debug.log"
		debugLogFile, err := os.OpenFile(debugLogFilePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			debugLogFile, _ = os.Create(debugLogFilePath)
		}
		gin.DefaultWriter = io.MultiWriter(debugLogFile, os.Stdout)

		errorLogFilePath := "log" + slash + "error.log"
		errLogFile, err := os.OpenFile(errorLogFilePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			errLogFile, _ = os.Create(errorLogFilePath)
		}
		gin.DefaultErrorWriter = io.MultiWriter(errLogFile)

		//log.SetOutput(gin.DefaultWriter)

		defer debugLogFile.Close()
		defer errLogFile.Close()
	*/

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

	//log.SetOutput(gin.DefaultWriter)

	d := &Daemon{r, config.GetConfig().Daemon.Listen}

	return d
}

func (d *Daemon) Run() {
	d.r.Run(d.listen)
}
