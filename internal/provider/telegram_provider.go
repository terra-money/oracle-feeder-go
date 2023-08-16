package provider

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramProvider struct {
	isSetup bool
	bot     *tgbotapi.BotAPI
	chatID  int64
}

func NewTelegramProvider() *TelegramProvider {
	var telegramBotKey string
	if telegramBotKey = os.Getenv("TELEGRAM_BOT_KEY"); telegramBotKey != "" {
		bot, err := tgbotapi.NewBotAPI(telegramBotKey)
		if err != nil {
			log.Fatal(err)
		}

		return &TelegramProvider{
			isSetup: true,
			bot:     bot,
			chatID:  -1001965303467,
		}
	}

	return &TelegramProvider{
		isSetup: false,
	}
}

func (p *TelegramProvider) SendError(info string, err error) error {
	if !p.isSetup {
		return nil
	}
	message := fmt.Sprintf("[ERROR][%s] %s", info, err)
	msg := tgbotapi.NewMessage(p.chatID, message)
	msg.ParseMode = "html"
	_, err = p.bot.Send(msg)
	return err
}

func (p *TelegramProvider) SendLog(message string) error {
	if !p.isSetup {
		return nil
	}
	msg := tgbotapi.NewMessage(p.chatID, message)
	msg.ParseMode = "html"
	_, err := p.bot.Send(msg)
	return err
}
