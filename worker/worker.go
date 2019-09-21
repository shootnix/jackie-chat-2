package worker

import (
	"fmt"
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
	log := logger.GetLogger()
	// Jackie Chat Daily chat id: -335599048
	u, err := entity.FindUser("Paolo")
	if err != nil {
		log.Fatal("Can't fild user for reporter: " + err.Error())
	}
	for _ = range time.Tick(w.timeInterval) {
		messagesSendTotal := entity.CountMessagesTotal(1)
		messagesSendToday := entity.CountMessagesToday(1)
		messagesFailToday := entity.CountMessagesToday(0)
		messagesFailTotal := entity.CountMessagesTotal(0)

		msg := fmt.Sprintf("Send Today: <b>%d</b>\nSend Total: <b>%d</b>\n", messagesSendToday, messagesSendTotal)
		msg = msg + fmt.Sprintf("Fail Today: <b>%d</b>\nFail Total: <b>%d</b>\n", messagesFailToday, messagesFailTotal)

		m := entity.NewMessage()
		m.Message = msg
		m.ChatID = -335599048
		m.BotID = 1
		m.UserID = u.ID
		m.ParseMode = "html"

		if err := m.Insert(); err != nil {
			log.Fatal("Can't insert message: " + err.Error())
		}
		m.Send("Telegram")
	}
}

// Перезабрасывает в очередь письма,
// которые не отправились почему-то
func (w *Checker) Run() {
	log := logger.GetLogger()
	sql := `

		SELECT id 
		  FROM messages
		 WHERE is_success = -1
		 ORDER BY ctime DESC

	`

	rows, err := io.GetPg().Conn.Query(sql)
	if err != nil {
		log.Error("Can't execute query " + sql + ": " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Error("Can't get message id: " + err.Error())
			continue
		}
		m, err := entity.GetMessage(id)
		if err != nil {
			log.Error("Can't get message by id: " + err.Error())
			continue
		}
		m.Send("telegram")
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
