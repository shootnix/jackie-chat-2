package main

import (
	"github.com/shootnix/jackie-chat-2/config"
	"github.com/shootnix/jackie-chat-2/daemon"
	"github.com/shootnix/jackie-chat-2/logger"
	"github.com/shootnix/jackie-chat-2/queue"
	"github.com/shootnix/jackie-chat-2/worker"
	"strconv"
	"time"
)

func main() {
	cfg := config.GetConfig()
	log := logger.GetLogger()
	queue.GetQueue()

	timeSum := 0
	nWorkers := 0
	for _, workerCfg := range cfg.Queue.Workers {
		if workerCfg.Type != "poster" {
			continue
		}
		timeSum = timeSum + workerCfg.TimeInterval
		nWorkers = nWorkers + 1
	}

	log.Debug("timeSum = " + strconv.FormatInt(int64(timeSum), 10))
	sleepInterval := time.Duration(int(timeSum/nWorkers)) * time.Second

	for _, workerCfg := range cfg.Queue.Workers {
		w := worker.NewWorker(workerCfg.Name, workerCfg.Type, time.Duration(workerCfg.TimeInterval)*time.Second)
		go w.Run()
		time.Sleep(sleepInterval)
	}

	daemon.NewDaemon().Run()
}
