package main

import (
	"log"
	"os"
	"strings"
	"strconv"
	"time"
	_ "time/tzdata"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	lastReplyTimeMap    map[int64]time.Time
	lastReminderTimeMap map[int64]time.Time
	lastUpdateTime      time.Time
	updateInterval  	= 2 * time.Hour
	checkInterval       = 5 * time.Minute
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
	if currentTime.Hour() >= 10 && currentTime.Hour() <= 20 {
		diff := currentTime.Sub(lastUpdateTime)
		if diff >= updateInterval {
			lastCheckTime, ok := lastReminderTimeMap[reminderChatID]
			if !ok || currentTime.Sub(lastCheckTime) >= reminderInterval {
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
}

func main() {
	phrases := []string{
		"Ах ты хитрая жопа",
		"Я все вижу",
		"Не так быстро, ковбой",
		"Я знаю все твои секреты",
		"Ты не сможешь уйти от меня",
		"Ты играешь с огнем",
		"Не подходи ближе",
		"Я буду следить за тобой",
		"Ты думал, что меня обманешь?",
		"Я знаю, что ты делал прошлым летом",
		"Ты в самой глубине моих мыслей",
		"Ничто не останется незамеченным",
		"Я всегда найду тебя",
		"Ты попался на мою удочку",
		"Я тебя предупреждал",
		"Не играй с огнем, малыш",
		"Ты в моей паутине",
		"Я знаю все твои слабости",
		"Ты не можешь скрыться от меня",
		"Я буду твоим худшим кошмаром",
		"Я следую за каждым твоим шагом",
		"Ты уже попался в сети",
		"Я знаю, кто ты на самом деле",
		"Не думай, что уйдешь от меня",
		"Я тебя найду, где бы ты ни был",
	}
	  	  
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Println(err)
	}
    time.Local = loc

	lastReplyTimeMap = make(map[int64]time.Time)
	lastReminderTimeMap = make(map[int64]time.Time)
	lastReminderTimeMap[reminderChatID] = time.Now()
	
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

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
					case "неа", "не-а", "no", "не":
						 // Инициализируем генератор случайных чисел
						r := rand.New(rand.NewSource(time.Now().UnixNano()))
						// Рандомно выбираем строки из списка
						selectedString := phrases[r.Intn(len(phrases))]
						reply := tgbotapi.NewMessage(chatID, selectedString)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					case "/get_id":
						chatIDStr := strconv.FormatInt(chatID, 10)
						reply := tgbotapi.NewMessage(chatID, chatIDStr)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
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
