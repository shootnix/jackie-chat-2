package entity

import (
	"errors"
	"github.com/shootnix/jackie-chat-2/io"
	"github.com/shootnix/jackie-chat-2/logger"
	"github.com/shootnix/jackie-chat-2/queue"
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
               user_id
          FROM messages
         WHERE id = $1

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
	if err := m.validate(); err != nil {
		return err
	}

	if m.ParseMode == "" {
		m.ParseMode = "html"
	}

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

func CountMessagesTotal(isSuccess int) int {

	sql := `

        SELECT COUNT(*) 
          FROM messages
         WHERE is_success = $1

    `

	var cnt int
	row := io.GetPg().Conn.QueryRow(sql, isSuccess)
	if err := row.Scan(&cnt); err != nil {
		return 0
	}

	return cnt
}

func CountMessagesToday(isSuccess int) int {

	sql := `

        SELECT COUNT(*) 
          FROM messages
         WHERE is_success = $1 
           AND ctime > now() - interval '1 day'

    `

	var cnt int
	row := io.GetPg().Conn.QueryRow(sql, isSuccess)
	if err := row.Scan(&cnt); err != nil {
		return 0
	}

	return cnt
}

func (m *Message) validate() error {
	if m.Message == "" {
		return errors.New("`Message` is required")
	}

	if m.ChatID == 0 {
		return errors.New("`ChatID` is required")
	}

	if m.BotID == 0 {
		return errors.New("`BotID` is required")
	}

	if m.UserID == 0 {
		return errors.New("`UserID` is required")
	}

	return nil
}

func (m *Message) Send(to string) error {
	log := logger.GetLogger()
	msgIDStr := strconv.FormatInt(m.ID, 10)
	log.Debug("Trying to send message #" + msgIDStr + " via " + to + "...")

	q := queue.GetQueue()
	q <- m.ID

	return nil
}
