package entity

import (
	"errors"
	"github.com/shootnix/jackie-chat-2/io"
	"github.com/shootnix/jackie-chat-2/logger"
	"strconv"
)

type Message struct {
	ID        int64  `db:"id"`
	Message   string `db:"message"`
	ParseMode string `db:"parse_mode"`
	ChatID    int64  `db:"chat_id"`
	CTime     string `db:"ctime"`
	IsSuccess int64  `db:"is_success"`
	Err       string `db:"err"`
	UserID    int64  `db:"user_id"`
	BotID     int64  `db:"bot_id"`
}

func NewMessage() *Message {
	m := new(Message)
	m.IsSuccess = -1
	m.Err = "waiting for delivery"

	return m
}

func GetMessage(id int64) (*Message, error) {
	m := new(Message)
	sql := `

        SELECT 
               id,
               message,
               parse_mode,
               chat_id,
               ctime,
               is_success,
               err,
               bot_id,
               user_id,
          FROM messages
         WHERE id = ?

    `
	row := io.GetPg().Conn.QueryRow(sql, id)
	err := row.Scan(
		&m.ID,
		&m.Message,
		&m.ParseMode,
		&m.ChatID,
		&m.CTime,
		&m.IsSuccess,
		&m.Err,
		&m.BotID,
		&m.UserID,
	)

	if err != nil {
		return m, err
	}

	return m, nil
}

func (m *Message) Update() error {
	if m.ID == 0 {
		return errors.New("ID of message is not defined!")
	}

	sql := `

        UPDATE messages
           SET message = $1,
               chat_id = $2,
               is_success = $3,
               err = $4
         WHERE id = $5

    `

	stmt, err := io.GetPg().Conn.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(m.Message, m.ChatID, m.IsSuccess, m.Err, m.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *Message) Insert() error {
	sql := `

        INSERT INTO messages (message, parse_mode, chat_id, bot_id, user_id, is_success, err)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id

    `
	row := io.GetPg().Conn.QueryRow(sql, m.Message, m.ParseMode, m.ChatID, m.BotID, m.UserID, m.IsSuccess, m.Err)
	if err := row.Scan(&m.ID); err != nil {
		return err
	}

	return nil
}

func (m *Message) Send(to string) error {
	log := logger.GetLogger()
	msgIDStr := strconv.FormatInt(m.ID, 10)
	log.Debug("Trying to send message #" + msgIDStr + " via " + to + "...")

	switch to {
	case "Telegram":
		if err := m.sendToTelegram(); err != nil {
			log.Error("Can't send message #" + msgIDStr + " via Telegram: " + err.Error())
			return err
		}
	default:
		return errors.New("Unknown transport: " + to)
	}

	return nil
}

func (m *Message) sendToTelegram() error {
	bot, err := GetTelegramBot(m.BotID)
	if err != nil {
		return err
	}

	tgm, err := io.NewTelegramBotAPI(bot.Token, m.Message, m.ChatID, m.ParseMode)
	if err != nil {
		return err
	}

	if err := tgm.SendMessage(); err != nil {
		return err
	}

	m.IsSuccess = 1
	m.Err = ""
	m.Update()

	log := logger.GetLogger()
	log.Debug("... Success! Message #" + strconv.FormatInt(m.ID, 10) + " has been sent via Telegram")

	return nil
}
