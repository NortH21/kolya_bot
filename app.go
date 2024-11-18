package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/anatoliyfedorenko/isdayoff"
)

var (
	lastReplyTimeMap    map[int64]time.Time
	lastReminderTimeMap map[int64]time.Time
	lastUpdateTime      time.Time
	updateInterval            = 2 * time.Hour
	checkInterval             = 1 * time.Minute
	reminderInterval          = 14 * time.Hour
	reminderChatID      int64 = -1002039497735
	// reminderChatID	int64 = 140450662
	// testId			int64 = -1001194083056
	meetUrl = "https://jitsi.sipleg.ru/spd"
)

var replyCountMap = make(map[int64]int)
const maxReplies = 3
const replyInterval = 35 * time.Minute

func isLastDayOfMonth(date time.Time) bool {
	nextDay := date.AddDate(0, 0, 1)
	return nextDay.Month() != date.Month()
}

func sendLastDayOfMonth(bot *tgbotapi.BotAPI) {
	reply := tgbotapi.NewMessage(reminderChatID, "📅 Сегодня последний день месяца! Напоминаю, что нужно заплатить за интернет.")
	_, err := bot.Send(reply)
	if err != nil {
		log.Println(err)
	}
}

func shouldSendReply(chatID int64) bool {
	currentTime := time.Now()
	lastReplyTime, exists := lastReplyTimeMap[chatID]
	replyCount, countExists := replyCountMap[chatID]

	if !exists || currentTime.Sub(lastReplyTime) >= replyInterval {
		lastReplyTimeMap[chatID] = currentTime
		replyCountMap[chatID] = 1
		return true
	} else if countExists && replyCount < maxReplies {
		replyCountMap[chatID]++
		return true
	}

	return false
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
	reminderMessage, err := getRandomLineFromFile("./files/reminder.txt")
	if err != nil {
		log.Fatal(err)
	}

	reply := tgbotapi.NewMessage(reminderChatID, reminderMessage)
	_, err = bot.Send(reply)
	if err != nil {
		log.Println(err)
	}
	lastReminderTimeMap[reminderChatID] = time.Now()
}

func getRandomLineFromFile(filename string) (string, error) {
	content, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer content.Close()

	scanner := bufio.NewScanner(content)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(lines))
	randomLine := lines[randomIndex]

	return randomLine, nil
}

func sendFridayGreetings(bot *tgbotapi.BotAPI) {
	fridayStr, err := getRandomLineFromFile("./files/friday.txt")
	if err != nil {
		log.Println("Ошибка при получении строки из файла:", err)
		return // Возвращаемся, чтобы избежать дальнейших действий при ошибке
	}

	reply := tgbotapi.NewMessage(reminderChatID, fridayStr)
	if _, err := bot.Send(reply); err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func sendMorningGreetings(bot *tgbotapi.BotAPI) {
	morningstr, err := getRandomLineFromFile("./files/morning.txt")
	if err != nil {
		log.Fatal(err)
	}

	morning := tgbotapi.NewMessage(reminderChatID, morningstr)
	_, err = bot.Send(morning)
	if err != nil {
		log.Println(err)
	}

	curTempYar, minTempYar, avgTempYar, maxTempYar, err := getTemperature("Yaroslavl")
	if err != nil {
		log.Println(err)
	}
	tempYar := fmt.Sprintf("В одном из старейших русских городов, основанном в XI веке и достигший своего расцвета в XVII веке, сейчас %d°C. Днем до %d°C, в среднем %d°C и ночью до %d°C.",
		curTempYar, maxTempYar, avgTempYar, minTempYar)

	curTempBak, minTempBak, avgTempBak, maxTempBak, err := getTemperature("Baku")
	if err != nil {
		log.Println(err)
	}
	tempBak := fmt.Sprintf("На Апшеронском полуострове в городе Бога сегодня тоже прекрасная погода, сейчас %d°C. Днем до %d°C, в среднем %d°C, ночью до %d°C.",
		curTempBak, maxTempBak, avgTempBak, minTempBak)

	rateUSD, err := getExchangeRates("USD")
	if err != nil {
		log.Println(err)
		return
	}

	rateAZN, err := getExchangeRates("AZN")
	if err != nil {
		log.Println(err)
		return
	}

	ratesUSDstr := fmt.Sprintf("Курс USD к рублю: %.2f.", rateUSD)
	ratesAZNstr := fmt.Sprintf("Курс AZN к рублю: %.2f.", rateAZN)
	ratesstr := fmt.Sprintf("%s \n%s", ratesUSDstr, ratesAZNstr)

	fullForecast := fmt.Sprintf("%s \n\n%s \n\n%s", tempYar, tempBak, ratesstr)
	forecast := tgbotapi.NewMessage(reminderChatID, fullForecast)
	_, err = bot.Send(forecast)
	if err != nil {
		log.Println(err)
	}
}

func main() {

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
			if update.ChannelPost != nil {
				channelMsg := update.ChannelPost
				if bot.Debug {
					log.Printf("Channel: [%s] %s", channelMsg.Chat.UserName, channelMsg.Text)
				}
			} else if update.Message != nil {
				if bot.Debug {
					log.Printf("Group: [%s] %s", update.Message.Chat.Title, update.Message.Text)
				}

				chatID := update.Message.Chat.ID
				replyToMessageID := update.Message.MessageID

				text := strings.ToLower(update.Message.Text)

				patternMeet := `(?i)(?:^|\s)(meet|мит|миит|миток)(?:$|\s)`
				reMeet := regexp.MustCompile(patternMeet)
				matchMeet := reMeet.MatchString(text)

				if matchMeet {
					text := ("Го, я создал " + meetUrl)
					reply := tgbotapi.NewMessage(chatID, text)
					if bot.Debug {
						log.Print(chatID, text)
					}

					_, err = bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
				}

				if strings.HasPrefix(text, "/chat") {
					commandText := strings.TrimSpace(strings.TrimPrefix(text, "/chat"))

					// typingMessage := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
					// if _, err := bot.Send(typingMessage); err != nil {
					// 	log.Println("Ошибка при отправке действия печати:", err)
					// }

					textResp := Chat(commandText)

					if textResp == "" {
						log.Println("Получен пустой текст от чата")
					} else {
						reply := tgbotapi.NewMessage(chatID, textResp)
						reply.ParseMode = tgbotapi.ModeMarkdown
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
					}
				}

				if strings.HasPrefix(text, "/img") {
					go func() {
						promt := strings.TrimSpace(strings.TrimPrefix(text, "/img"))

						if promt == "" {
							log.Println("Нет промта, ничего не делаем")
							reply := tgbotapi.NewMessage(chatID, "Нужно указать промт")
							reply.ReplyToMessageID = replyToMessageID
							_, err := bot.Send(reply)
							if err != nil {
								log.Println(err)
							}
						} else {
							negativepromt := ""

							fileName, err := getImage(promt, negativepromt)
							if err != nil {
								log.Println(err)
							}

							if fileName == "" {
								log.Println("Картинка не вернулась")
							} else {
								photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(fileName))
								if _, err = bot.Send(photo); err != nil {
									log.Println(err)
								}
							}
						}
					}()
				}

				usernameWithAt := strings.ToLower("@" + bot.Self.UserName)

				rand.Seed(time.Now().UnixNano())
				switch text {
				case "да", "да)", "да!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Пизда")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "мда", "мда)", "мда!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Манда")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "нет", "нет)", "нет!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Пидора ответ")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "неа", "не-а", "no", "не", "неа)", "не)", "отнюдь":
					if shouldSendReply(chatID) {
						nostr, err := getRandomLineFromFile("./files/no.txt")
						if err != nil {
							log.Fatal(err)
						}
						reply := tgbotapi.NewMessage(chatID, nostr)
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err = bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "a", "а", "a)", "а)", "а!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Хуй на)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "естественно", "естественно)", "естественно!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Хуестественно)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "конечно", "конечно)", "конечно!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Хуечно")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "300", "триста", "тристо", "три сотни", "3 сотки", "три сотки":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Отсоси у тракториста)))")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "как сам", "как сам?":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Как сало килограмм")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "именно", "именно)", "именно!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Хуименно")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "хуй на":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "А тебе два)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "ну вот":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Хуй тебе в рот)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "нет, тебе", "нет тебе", "нет, тебе!", "нет тебе!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Нет, тебе!)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "нет, ты", "нет ты", "нет, ты!", "нет ты!":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Нет, ты!)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					}
				case "пинг", "ping", "зштп", "gbyu":
					if shouldSendReply(chatID) {
						reply := tgbotapi.NewMessage(chatID, "Хуинг")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
					}
				case "/get_id", "/get_id" + usernameWithAt:
					chatIDStr := strconv.FormatInt(chatID, 10)
					reply := tgbotapi.NewMessage(chatID, chatIDStr)
					_, err := bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
				case "/forecast", "/forecast" + usernameWithAt:
					curTempYar, minTempYar, avgTempYar, maxTempYar, err := getTemperature("Yaroslavl")
					if err != nil {
						log.Println(err)
					}
					tempYar := fmt.Sprintf("В Ярославле сейчас %d°C. Днем до %d°C, в среднем %d°C и ночью до %d°C.",
						curTempYar, maxTempYar, avgTempYar, minTempYar)

					curTempBak, minTempBak, avgTempBak, maxTempBak, err := getTemperature("Baku")
					if err != nil {
						log.Println(err)
					}
					tempBak := fmt.Sprintf("В Баку сейчас %d°C. Днем до %d°C, в среднем %d°C, ночью до %d°C.",
						curTempBak, maxTempBak, avgTempBak, minTempBak)

					fullForecast := fmt.Sprintf("%s \n\n%s", tempYar, tempBak)

					reply := tgbotapi.NewMessage(chatID, fullForecast)
					_, err = bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
				case "/rates", "/rates" + usernameWithAt:
					rateUSD, err := getExchangeRates("USD")
					if err != nil {
						fmt.Println(err)
						return
					}

					rateAZN, err := getExchangeRates("AZN")
					if err != nil {
						fmt.Println(err)
						return
					}

					ratesUSDstr := fmt.Sprintf("Курс USD к рублю: %.2f.", rateUSD)
					ratesAZNstr := fmt.Sprintf("Курс AZN к рублю: %.2f.", rateAZN)
					ratesstr := fmt.Sprintf("%s \n%s", ratesUSDstr, ratesAZNstr)
					reply := tgbotapi.NewMessage(chatID, ratesstr)
					_, err = bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
				case "/jokes", "/jokes" + usernameWithAt:
					text, err := getJokes()
					if err != nil {
						log.Fatal(err)
					}
					if text == "" {
						log.Println("Получен пустой текст шутки, выполнение прекращено.")
						return
					}
					reply := tgbotapi.NewMessage(chatID, text)
					_, err = bot.Send(reply)
					if err != nil {
						log.Println(err)
					}
				case usernameWithAt:
					// Список статей
					ukrf, err := getRandomLineFromFile("./files/ukrf.txt")
					if err != nil {
						log.Fatal(err)
					}
					reply := tgbotapi.NewMessage(chatID, ukrf)
					if bot.Debug {
						log.Print(chatID, ukrf)
					}
					_, err = bot.Send(reply)
					if err != nil {
						log.Println(err)
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

	// Friday/morning loop
	go func() {
		for {
			currentTime := time.Now()
			if (currentTime.Month() >= time.April && currentTime.Month() <= time.August && currentTime.Hour() == 7 && currentTime.Minute() == 0) ||
				(currentTime.Month() < time.April || currentTime.Month() > time.August && currentTime.Hour() == 8 && currentTime.Minute() == 0) {
				go sendMorningGreetings(bot)
			}

			workdayInfo, err := CheckWorkday()
			if err != nil {
				continue
			}

			if workdayInfo.Today != nil && *workdayInfo.Today == isdayoff.DayTypeWorking {
				tomorrow := workdayInfo.Tomorrow
				if tomorrow != nil && *tomorrow == isdayoff.DayTypeNonWorking {
					if currentTime.Hour() == 17 && currentTime.Minute() == 0 {
						go sendFridayGreetings(bot)
					}
				}
			}
			time.Sleep(checkInterval)
		}
	}()

	// Last day loop
	go func() {
		for {
			currentTime := time.Now()
			if isLastDayOfMonth(currentTime) && currentTime.Hour() == 12 && currentTime.Minute() == 0 {
				sendLastDayOfMonth(bot)
			}
			time.Sleep(checkInterval)
		}
	}()

	// Keep main goroutine alive
	select {}
}
