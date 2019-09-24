package worker

import (
	"github.com/shootnix/jackie-chat-2/logger"
	"time"
)

type Worker interface {
	Run()
}

func NewWorker(wName, wType string, timeInt time.Duration) Worker {
	log := logger.GetLogger()

	log.Debug("worker type " + wType)

	switch wType {
	case "reporter":
		return NewReporter(wName, timeInt)
	case "poster":
		return NewPoster(wName, timeInt)
	case "checker":
		return NewChecker(wName, timeInt)
	default:
		log.Fatal("Unknown worker type: " + wType)
		return nil
	}
}
