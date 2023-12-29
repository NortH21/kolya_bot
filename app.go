package main

import (
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Переменная для хранения времени последнего сообщения
	var lastMessageTime time.Time

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"channel_post", "message"}

	updates := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.ChannelPost != nil {
			// Обработка постов в каналах
		} else if update.Message != nil {
			// Обработка сообщений в группах/чатах
			// Преобразование текста сообщения в нижний регистр для сравнения
			lowerText := strings.ToLower(update.Message.Text)

			// Проверка, что "да" встречается в конце предложения
			if strings.HasSuffix(lowerText, "да") {
				reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Пизда")
				_, err := bot.Send(reply)
				if err != nil {
					log.Println(err)
				}

				// Обновление времени последнего сообщения
				lastMessageTime = time.Now()
			} else if strings.HasSuffix(lowerText, "нет") {
				reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Пидора ответ")
				_, err := bot.Send(reply)
				if err != nil {
					log.Println(err)
				}

				// Обновление времени последнего сообщения
				lastMessageTime = time.Now()
			}
		}

		// Проверка на отправку сообщения "Есть живые тут?" после 6 часов без активности
		if time.Since(lastMessageTime).Hours() >= 6 {
			// Проверка, что текущее время не в интервале 23:00 - 09:00 по Москве
			loc, _ := time.LoadLocation("Europe/Moscow")
			currentTime := time.Now().In(loc)

			if currentTime.Hour() >= 9 && currentTime.Hour() < 23 {
				// Отправка сообщения
				reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Есть живые тут? Как дела в целом?")
				_, err := bot.Send(reply)
				if err != nil {
					log.Println(err)
				}

				// Обновление времени последнего сообщения
				lastMessageTime = time.Now()
			}
		}
	}
}
