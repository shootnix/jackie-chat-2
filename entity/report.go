package entity

import (
	"fmt"
	"github.com/shootnix/jackie-chat-2/constant"
	"github.com/shootnix/jackie-chat-2/logger"
)

type Report struct {
	Text string
}

func NewReport() *Report {
	r := new(Report)

	return r
}

func (r *Report) AppendMessagesTotal(isSuccess int) {
	total := CountMessagesTotal(isSuccess)
	text := fmt.Sprintf("Send Total: <b>%d</b>\n", total)
	r.Text = r.Text + text
}

func (r *Report) AppendMessagesToday(isSuccess int) {
	total := CountMessagesToday(isSuccess)
	text := fmt.Sprintf("Send Today: <b>%d</b>\n", total)
	r.Text = r.Text + text
}

func (r *Report) Send() error {
	log := logger.GetLogger()
	u, err := FindUser("Paolo")
	if err != nil {
		//log.Fatal("Can't fild user for reporter: " + err.Error())
		return err
	}

	m := NewMessage()
	m.Message = r.Text
	m.ChatID = constant.JACKIE_CHAT_DAILY
	m.BotID = 1
	m.UserID = u.ID
	m.ParseMode = "html"

	if err := m.Insert(); err != nil {
		log.Fatal("Can't insert message: " + err.Error())
	}
	m.Send("Telegram")

	return nil
}
