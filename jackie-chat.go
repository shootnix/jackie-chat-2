package main

import (
	//"github.com/shootnix/jackie-chat-2/config"
	"github.com/shootnix/jackie-chat-2/daemon"
	//"github.com/shootnix/jackie-chat-2/io"
	"github.com/shootnix/jackie-chat-2/logger"
)

func main() {
	//conf := config.GetConfig()
	//pg := io.GetPg()
	//defer pg.Conn.Close()

	log := logger.GetLogger()

	//log.Debug(conf.Database.Pg.DBName)
	//log.Error("Oops...")

	log.Debug("======>>>>>")
	log.Info(">>>>>>>>>>>")

	daemon.NewDaemon().Run()
}
