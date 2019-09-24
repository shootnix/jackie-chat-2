package worker

import (
	"github.com/shootnix/jackie-chat-2/entity"
	"time"
)

type Checker struct {
	timeInterval time.Duration
	Name         string
}

func NewChecker(name string, timeInt time.Duration) *Checker {
	w := new(Checker)
	w.Name = name
	w.timeInterval = timeInt

	return w
}

// Перезабрасывает в очередь письма,
// которые не отправились почему-то
func (w *Checker) Run() {
	for _ = range time.Tick(w.timeInterval) {
		pendingMessages := entity.ListPendingMessages()
		for _, m := range pendingMessages {
			m.Send("telegram")
		}
	}
}
