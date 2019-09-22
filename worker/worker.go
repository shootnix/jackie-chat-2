package worker

import (
	//"fmt"
	//"github.com/shootnix/jackie-chat-2/constant"
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

type Reporter struct {
	timeInterval time.Duration
	Name         string
}

type Checker struct {
	timeInterval time.Duration
	Name         string
}

type Worker interface {
	Run()
}

func NewPoster(name string, timeInt time.Duration) *Poster {
	w := new(Poster)
	w.Name = name
	w.timeInterval = timeInt

	return w
}

func NewReporter(name string, timeInt time.Duration) *Reporter {
	w := new(Reporter)
	w.Name = name
	w.timeInterval = timeInt

	return w
}

func NewChecker(name string, timeInt time.Duration) *Checker {
	w := new(Checker)
	w.Name = name
	w.timeInterval = timeInt

	return w
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

func (w *Poster) Run() {
	q := queue.GetQueue()
	for _ = range time.Tick(w.timeInterval) {
		if ID := <-q; ID != 0 {
			// if got message in the queue, try to send it:
			go w.sendTgmMessage(ID)
		}
	}
}

func (w *Reporter) Run() {
	for _ = range time.Tick(w.timeInterval) {
		report := entity.NewReport()
		report.AppendMessagesTotal()
		report.AppendMessagesToday()
		report.Send()
	}
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

	j := entity.NewJournal(w.Name, m.ID)
	j.Insert()
}
