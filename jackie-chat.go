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
	cfg := config.GetConfig()
	log := logger.GetLogger()
	queue.GetQueue()
	sleepInterval := time.Duration(int(cfg.Queue.WorkerTimeInterval/2)) * time.Second

	for _, workerName := range cfg.Queue.Workers {
		w := worker.NewWorker(time.Duration(cfg.Queue.WorkerTimeInterval)*time.Second, workerName)
		log.Debug("Starting worker " + w.Name)
		go w.Run()
		time.Sleep(sleepInterval)
	}

	daemon.NewDaemon().Run()
}
