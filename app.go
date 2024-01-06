package main

import (
	"log"
	"os"
	"strings"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	lastReplyTimeMap    map[int64]time.Time
	lastReminderTimeMap map[int64]time.Time
	lastUpdateTime      time.Time
	lastUpdateInterval  = 1 * time.Hour
	checkInterval       = 1 * time.Minute
	reminderInterval    = 24 * time.Hour
	reminderMessage     = "Ну чо, посоны, вы как? Живы?"
	reminderChatID      int64 = -1002039497735
)

func shouldSendReply(chatID int64) bool {
	currentTime := time.Now()
	diff := currentTime.Sub(lastReplyTimeMap[chatID])
	return diff.Minutes() >= 25
}

func shouldSendReminder() bool {
	currentTime := time.Now()
	if currentTime.Hour() >= 9 && currentTime.Hour() <= 21 {
		lastCheckTime, ok := lastReminderTimeMap[reminderChatID]
		if !ok || currentTime.Sub(lastCheckTime) >= reminderInterval {
			diff := currentTime.Sub(lastUpdateTime)
			if diff >= lastUpdateInterval {
				return true
			}
		}
	}
	return false
}

func sendReminder(bot *tgbotapi.BotAPI) {
	reply := tgbotapi.NewMessage(reminderChatID, reminderMessage)
	_, err := bot.Send(reply)
	if err != nil {
		log.Println(err)
	}
	lastReminderTimeMap[reminderChatID] = time.Now()
	log.Print("lastReminderTime: ", time.Now())
}

func main() {
	lastReplyTimeMap = make(map[int64]time.Time)
	lastReminderTimeMap = make(map[int64]time.Time)
	lastReminderTimeMap[reminderChatID] = time.Now()
	
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

	// Updates loop
	go func() {
		for update := range updates {
			lastUpdateTime = time.Now()
			log.Print(lastUpdateTime)
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
					text := strings.ToLower(update.Message.Text)
					switch text {
					case "да":
						reply := tgbotapi.NewMessage(chatID, "Пизда")
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					case "нет":
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
	}()

	// Reminder loop
	go func() {
		for {
			if shouldSendReminder() {
				sendReminder(bot)
			}
			time.Sleep(checkInterval)
		}
	}()

	// Keep main goroutine alive
	select {}
}
