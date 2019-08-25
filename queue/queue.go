package queue

import (
	"github.com/shootnix/jackie-chat-2/config"
	"sync"
)

var once sync.Once
var q chan int64

func GetQueue() chan int64 {
	once.Do(func() {
		cfg := config.GetConfig()
		q = make(chan int64, cfg.Queue.Size)
	})
	return q
}
