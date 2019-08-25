package entity

import (
	"github.com/shootnix/jackie-chat-2/io"
)

type TelegramBot struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	Token    string `db:"token"`
	IsActive bool   `db:"is_active"`
	Info     string `db:"info"`
}

func GetTelegramBot(id int64) (*TelegramBot, error) {
	sql := `

        SELECT id, name, token
          FROM telegram_bots
         WHERE id = $1
           AND is_active = true

    `

	bot := new(TelegramBot)
	db := io.GetPg().Conn
	row := db.QueryRow(sql, id)
	err := row.Scan(&bot.ID, &bot.Name, &bot.Token)
	if err != nil {
		return bot, err
	}

	return bot, nil
}
