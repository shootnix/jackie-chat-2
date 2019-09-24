package worker

import (
	"github.com/shootnix/jackie-chat-2/entity"
	"time"
)

type Reporter struct {
	timeInterval time.Duration
	Name         string
}

func NewReporter(name string, timeInt time.Duration) *Reporter {
	w := new(Reporter)
	w.Name = name
	w.timeInterval = timeInt

	return w
}

func (w *Reporter) Run() {
	for _ = range time.Tick(w.timeInterval) {
		report := entity.NewReport()
		report.AppendMessagesTotal()
		report.AppendMessagesToday()
		report.Send()
	}
}
