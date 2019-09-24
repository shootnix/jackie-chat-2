package worker

import (
	"github.com/shootnix/jackie-chat-2/entity"
	"github.com/shootnix/jackie-chat-2/io"
	"github.com/shootnix/jackie-chat-2/logger"
	"github.com/shootnix/jackie-chat-2/queue"
	"strconv"
	"time"
)

type Poster struct {
	timeInterval time.Duration
	Name         string
}

func NewPoster(name string, timeInt time.Duration) *Poster {
	w := new(Poster)
	w.Name = name
	w.timeInterval = timeInt

	return w
}

func (w *Poster) Run() {
	q := queue.GetQueue()
	for _ = range time.Tick(w.timeInterval) {
		if ID := <-q; ID != 0 {
			// if got message in the queue, try to send it:
			go w.sendTgmMessage(ID)
		}
	}
}

func (w *Poster) sendTgmMessage(id int64) {
	log := logger.GetLogger()
	log.Debug(w.Name + ": SENDING MESSAGE: " + strconv.FormatInt(id, 10))
	m, err := entity.GetMessage(id)
	if err != nil {
		log.Error("Can't get message: " + err.Error())
		return
	}

	bot, err := entity.GetTelegramBot(m.BotID)
	if err != nil {
		errMsg := "Can't get telegram bot: " + err.Error()
		log.Error(errMsg)

		m.Err = errMsg
		m.IsSuccess = 0
		m.Update()

		return
	}

	tgm, err := io.NewTelegramBotAPI(bot.Token, m.Message, m.ChatID, m.ParseMode)
	if err != nil {
		errMsg := "Can't create a telegram bot api: " + err.Error()
		log.Error(errMsg)

		m.Err = errMsg
		m.IsSuccess = 0
		m.Update()

		return
	}

	if err := tgm.SendMessage(); err != nil {
		errMsg := "Can't send a message: " + err.Error()
		log.Error(errMsg)

		m.Err = errMsg
		m.IsSuccess = 0
		m.Update()

		return
	}

	m.IsSuccess = 1
	m.Err = ""
	m.Update()

	j := entity.NewJournal(w.Name, m.ID)
	j.Insert()
}
