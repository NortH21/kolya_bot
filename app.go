package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	lastReplyTimeMap    map[int64]time.Time
	lastReminderTimeMap map[int64]time.Time
	lastUpdateTime      time.Time
	updateInterval            = 2 * time.Hour
	checkInterval             = 1 * time.Minute
	reminderInterval          = 24 * time.Hour
	reminderChatID      int64 = -1002039497735
	//reminderChatIDTest	int64 = 140450662
	//testId			int64 = -1001194083056
	meetUrl = "https://jitsi.sipleg.ru/spd"
)

func getTemperature(city string) (int, int, int, int, error) {
	type Forecast struct {
		Forecastday []struct {
			Day struct {
				MaxTempC float64 `json:"maxtemp_c"`
				MinTempC float64 `json:"mintemp_c"`
				AvgTempC float64 `json:"avgtemp_c"`
			} `json:"day"`
		} `json:"forecastday"`
	}

	type Current struct {
		TempC float64 `json:"temp_c"`
	}

	type WeatherData struct {
		Current  Current  `json:"current"`
		Forecast Forecast `json:"forecast"`
	}

	url := "https://api.weatherapi.com/v1/forecast.json?q=" + city + "&days=1&key=" + os.Getenv("weatherapi_key")

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	maxTemp := weatherData.Forecast.Forecastday[0].Day.MaxTempC
	minTemp := weatherData.Forecast.Forecastday[0].Day.MinTempC
	avgTemp := weatherData.Forecast.Forecastday[0].Day.AvgTempC
	curTemp := weatherData.Current.TempC

	return int(curTemp), int(minTemp), int(avgTemp), int(maxTemp), err
}

func shouldSendReply(chatID int64) bool {
	currentTime := time.Now()
	diff := currentTime.Sub(lastReplyTimeMap[chatID])
	return diff.Minutes() >= 35
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
	fridaystr, err := getRandomLineFromFile("./files/friday.txt")
	if err != nil {
		log.Fatal(err)
	}

	reply := tgbotapi.NewMessage(reminderChatID, fridaystr)
	_, err = bot.Send(reply)
	if err != nil {
		log.Println(err)
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
	tempYar := fmt.Sprintf("В одном из старейших русских городов, основанный в XI веке и достигший своего расцвета в XVII веке, сейчас %d°C. Сегодня тепература будет до %d°C, в среднем %d°C и минимально до %d°C.",
		curTempYar, maxTempYar, avgTempYar, minTempYar)

	curTempBak, minTempBak, avgTempBak, maxTempBak, err := getTemperature("Baku")
	if err != nil {
		log.Println(err)
	}
	tempBak := fmt.Sprintf("На Апшеронском полуострове в городе Бога сегодня тоже прекрасная погода, сейчас %d°C. Днем до %d°C, в среднем %d°C, ночью до %d°C.",
		curTempBak, maxTempBak, avgTempBak, minTempBak)

	fullForecast := fmt.Sprintf("%s \n\n%s", tempYar, tempBak)
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

				patternMeet := `/meet|мит|миит|миток/i`
				reMeet := regexp.MustCompile(patternMeet)
				matchMeet := reMeet.MatchString(text)

				if matchMeet && meetUrl != "" {
					if update.Message.From.ID == 113501382 {
						reply := tgbotapi.NewMessage(chatID, "Бебебе мемеме")

						_, err = bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
					} else {
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
				}

				if shouldSendReply(chatID) {
					usernameWithAt := strings.ToLower("@" + bot.Self.UserName)

					rand.Seed(time.Now().UnixNano())
					switch text {
					case "да":
						reply := tgbotapi.NewMessage(chatID, "Пизда")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					case "нет":
						reply := tgbotapi.NewMessage(chatID, "Пидора ответ")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
						_, err := bot.Send(reply)
						if err != nil {
							log.Println(err)
						}
						lastReplyTimeMap[chatID] = time.Now()
					case "неа", "не-а", "no", "не":
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
					case "a", "а":
						reply := tgbotapi.NewMessage(chatID, "Хуй на)")
						reply.ReplyToMessageID = replyToMessageID
						time.Sleep(2 * time.Second)
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
					case "/forecast":
						curTempYar, minTempYar, avgTempYar, maxTempYar, err := getTemperature("Yaroslavl")
						if err != nil {
							log.Println(err)
						}
						tempYar := fmt.Sprintf("В одном из старейших русских городов, основанный в XI веке и достигший своего расцвета в XVII веке, сейчас %d°C. Сегодня тепература будет до %d°C, в среднем %d°C и минимально до %d°C.",
							curTempYar, maxTempYar, avgTempYar, minTempYar)

						curTempBak, minTempBak, avgTempBak, maxTempBak, err := getTemperature("Baku")
						if err != nil {
							log.Println(err)
						}
						tempBak := fmt.Sprintf("На Апшеронском полуострове в городе Бога сегодня тоже прекрасная погода, сейчас %d°C. Днем до %d°C, в среднем %d°C, ночью до %d°C.",
							curTempBak, maxTempBak, avgTempBak, minTempBak)
					
						fullForecast := fmt.Sprintf("%s \n\n%s", tempYar, tempBak)

						reply := tgbotapi.NewMessage(chatID, fullForecast)
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
			if currentTime.Weekday() == time.Friday && currentTime.Hour() == 17 && currentTime.Minute() == 0 {
				sendFridayGreetings(bot)
			}
			if currentTime.Hour() == 8 && currentTime.Minute() == 0 {
				sendMorningGreetings(bot)
			}
			time.Sleep(checkInterval)
		}
	}()

	// Keep main goroutine alive
	select {}
}
