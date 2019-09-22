package io

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBotAPI struct {
	Bot *tgbotapi.BotAPI
	Msg tgbotapi.MessageConfig
}

func NewTelegramBotAPI(token, message string, chatID int64, parseMode string) (*TelegramBotAPI, error) {
	tgm := new(TelegramBotAPI)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return tgm, err
	}
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = parseMode

	tgm.Bot = bot
	tgm.Msg = msg

	return tgm, nil
}

func (tgm *TelegramBotAPI) SendMessage() error {
	if _, err := tgm.Bot.Send(tgm.Msg); err != nil {
		return err
	}

	return nil
}
