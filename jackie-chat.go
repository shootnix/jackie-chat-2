package main

import (
	"github.com/shootnix/jackie-chat-2/config"
	"github.com/shootnix/jackie-chat-2/daemon"
	"github.com/shootnix/jackie-chat-2/logger"
	"github.com/shootnix/jackie-chat-2/queue"
	"github.com/shootnix/jackie-chat-2/worker"
	"time"
)

func main() {
	_ = config.GetConfig()

	logger.GetLogger()
	queue.GetQueue()

	worker := worker.NewWorker(1 * time.Second)
	go worker.Run()

	daemon.NewDaemon().Run()
}
