package worker

import (
	"github.com/shootnix/jackie-chat-2/entity"
	"github.com/shootnix/jackie-chat-2/io"
	"github.com/shootnix/jackie-chat-2/logger"
	"github.com/shootnix/jackie-chat-2/queue"
	"strconv"
	"time"
)

type Worker struct {
	timeInterval time.Duration
}

func NewWorker(timeInt time.Duration) *Worker {
	w := new(Worker)
	w.timeInterval = timeInt

	return w
}

func (w *Worker) Run() {
	q := queue.GetQueue()
	for _ = range time.Tick(w.timeInterval) {
		if ID := <-q; ID != 0 {
			// if got message in the queue, try to send it:
			go w.sendTgmMessage(ID)
		}
	}
}

func (w *Worker) sendTgmMessage(id int64) {
	log := logger.GetLogger()
	log.Debug("SEND MESSAGE: " + strconv.FormatInt(id, 10))
	m, err := entity.GetMessage(id)
	if err != nil {
		log.Error("Can't get message: " + err.Error())
		return
	}

	bot, err := entity.GetTelegramBot(m.BotID)
	if err != nil {
		log.Error("Can't get telegram bot: " + err.Error())
	}

	tgm, err := io.NewTelegramBotAPI(bot.Token, m.Message, m.ChatID, m.ParseMode)
	if err != nil {
		log.Error("Can't create a telegram bot api: " + err.Error())
		return
	}

	if err := tgm.SendMessage(); err != nil {
		log.Error("Can't send a message: " + err.Error())
		m.IsSuccess = 0
		m.Update()

		return
	}

	m.IsSuccess = 1
	m.Err = ""
	m.Update()
}
