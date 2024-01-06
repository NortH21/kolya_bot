package main

import (
	"os"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var lastReplyTimeMap map[int64]time.Time

func shouldSendReply(chatID int64) bool {
	currentTime := time.Now()
	diff := currentTime.Sub(lastReplyTimeMap[chatID])
	return diff.Minutes() >= 15
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"channel_post", "message"}

	updates := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	lastReplyTimeMap = make(map[int64]time.Time)

	for update := range updates {
		if update.ChannelPost != nil {
			channelMsg := update.ChannelPost
			if bot.Debug {
				log.Printf("Channel: [%s] %s", channelMsg.Chat.UserName, channelMsg.Text)
			}
		} else if update.Message != nil {
			groupMsg := update.Message
			if bot.Debug {
				log.Printf("Group: [%s] %s", groupMsg.Chat.UserName, groupMsg.Text)
			}

			chatID := update.Message.Chat.ID

			if shouldSendReply(chatID) {
				if strings.ToLower(update.Message.Text) == "да" {
					reply := tgbotapi.NewMessage(chatID, "Пизда")
					_, err := bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
					lastReplyTimeMap[chatID] = time.Now()
				} else if strings.ToLower(update.Message.Text) == "нет" {
					reply := tgbotapi.NewMessage(chatID, "Пидора ответ")
					_, err := bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
					lastReplyTimeMap[chatID] = time.Now()
				}
			}
		}
	}
}

