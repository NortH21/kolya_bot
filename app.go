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

	"github.com/anatoliyfedorenko/isdayoff"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	fridayStr := Chat("поздравь коллег с окончанием рабочей недели и добавь смайлики, без особого формализма")
	if fridayStr == "" {
		log.Println("Получен пустой текст от чата")
		
		var err error
		fridayStr, err = getRandomLineFromFile("./files/friday.txt")
		if err != nil {
			log.Println("Ошибка при получении строки из файла:", err)
			return
		}
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

func sendReply(bot *tgbotapi.BotAPI, chatID int64, replyToMessageID int, text string) {
	if shouldSendReply(chatID) {
		reply := tgbotapi.NewMessage(chatID, text)
		reply.ReplyToMessageID = replyToMessageID
		time.Sleep(2 * time.Second)
		_, err := bot.Send(reply)
		if err != nil {
			log.Println(err)
		}
		lastReplyTimeMap[chatID] = time.Now()
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

				// typingMessage := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
				// bot.Send(typingMessage)

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
								// typingMessage := tgbotapi.NewChatAction(chatID, tgbotapi.ChatUploadPhoto)
								// bot.Send(typingMessage)
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
					sendReply(bot, chatID, update.Message.MessageID, "Пизда")
				case "мда", "мда)", "мда!":
					sendReply(bot, chatID, update.Message.MessageID, "Манда")
				case "нет", "нет)", "нет!":
					sendReply(bot, chatID, update.Message.MessageID, "Пидора ответ")
				case "a", "а", "a)", "а)", "а!":
					sendReply(bot, chatID, update.Message.MessageID, "Хуй на)")
				case "естественно", "естественно)", "естественно!":
					sendReply(bot, chatID, update.Message.MessageID, "Хуестественно)")
				case "чо", "чо?", "чо?)":
					sendReply(bot, chatID, update.Message.MessageID, "Хуй в очо)")
				case "конечно", "конечно)", "конечно!":
					sendReply(bot, chatID, update.Message.MessageID, "Хуечно)")
				case "300", "триста", "тристо", "три сотни", "3 сотки", "три сотки":
					sendReply(bot, chatID, update.Message.MessageID, "Отсоси у тракториста)))")
				case "как сам", "как сам?":
					sendReply(bot, chatID, update.Message.MessageID, "Как сало килограмм")
				case "именно", "именно)", "именно!":
					sendReply(bot, chatID, update.Message.MessageID, "Хуименно")
				case "хуй на":
					sendReply(bot, chatID, update.Message.MessageID, "А тебе два)")
				case "ну вот":
					sendReply(bot, chatID, update.Message.MessageID, "Хуй тебе в рот)")
				case "нет, тебе", "нет тебе", "нет, тебе!", "нет тебе!":
					sendReply(bot, chatID, update.Message.MessageID, "Нет, тебе!)")
				case "нет, ты", "нет ты", "нет, ты!", "нет ты!":
					sendReply(bot, chatID, update.Message.MessageID, "Нет, ты!)")
				case "пинг", "ping", "зштп", "gbyu":
					sendReply(bot, chatID, update.Message.MessageID, "Хуинг")
				case "/get_id", "/get_id" + usernameWithAt:
					chatIDStr := strconv.FormatInt(chatID, 10)
					sendReply(bot, chatID, update.Message.MessageID, chatIDStr)
				case "неа", "не-а", "no", "не", "неа)", "не)", "отнюдь":
					nostr, err := getRandomLineFromFile("./files/no.txt")
					if err != nil {
						log.Fatal(err)
					}
					sendReply(bot, chatID, update.Message.MessageID, nostr)
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
			log.Println("currentTime: ", currentTime, "currentTime.Month(): ", currentTime.Month(), "currentTime.Hour():", currentTime.Hour())
			// if (currentTime.Month() >= time.April && currentTime.Month() <= time.August && currentTime.Hour() == 7 && currentTime.Minute() == 0) ||
			// 	(currentTime.Month() < time.April || currentTime.Month() > time.August && currentTime.Hour() == 8 && currentTime.Minute() == 0) {
			// 	go sendMorningGreetings(bot)
			// }

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
