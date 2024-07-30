package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
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
	reminderInterval          = 14 * time.Hour
	reminderChatID      int64 = -1002039497735
	rChatID 			int64 = 113501382
	rInterval   			  = 1 * time.Hour
	//reminderChatIDTest	int64 = 140450662
	//testId			int64 = -1001194083056
	meetUrl = "https://jitsi.sipleg.ru/spd"
)

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

func generateJokesURL(pid, key string) string {
	uts := strconv.FormatInt(time.Now().Unix(), 10)
	query := url.Values{}
	query.Set("pid", pid)
	query.Set("method", "getRandItem")
	query.Set("uts", uts)
	query.Set("category", "4") // 4 – чёрный юмор
	query.Set("genre", "1")    // 1 – анекдоты

	hash := md5.Sum([]byte(query.Encode() + key))

	u := url.URL{
		Scheme:   "http",
		Host:     "anecdotica.ru",
		Path:     "/api",
		RawQuery: query.Encode() + "&hash=" + fmt.Sprintf("%x", hash),
	}

	return u.String()
}

func getJokes() (string, error) {
	type AnecdoteResponse struct {
		Result struct {
			Error  int    `json:"error"`
			ErrMsg string `json:"errMsg"`
		} `json:"result"`
		Item struct {
			Text string `json:"text"`
			Note string `json:"note"`
		} `json:"item"`
	}

	pid := os.Getenv("anecdotica_pid")
	key := os.Getenv("anecdotica_key")

	url := generateJokesURL(pid, key)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	var anecdoteResponse AnecdoteResponse
	err = json.Unmarshal(body, &anecdoteResponse)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", err
	}

	if anecdoteResponse.Result.Error == 0 {
		fmt.Println("Anecdote:", anecdoteResponse.Item.Text)
	} else {
		fmt.Println("Error:", anecdoteResponse.Result.ErrMsg)
	}
	return anecdoteResponse.Item.Text, err
}

func getExchangeRates(currencyCode string) (float64, error) {
	resp, err := http.Get("https://www.cbr-xml-daily.ru/daily_json.js")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	if valute, ok := data["Valute"].(map[string]interface{}); ok {
		if currency, ok := valute[currencyCode].(map[string]interface{}); ok {
			rate := currency["Value"].(float64)
			return rate, nil
		} else {
			return 0, err
		}
	}
	return 0, err
}

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

	if city == "Yaroslavl" {
		return int(curTemp) + 2, int(minTemp) + 2, int(avgTemp) + 2, int(maxTemp) + 3, err // ну бля
	} else {
		return int(curTemp), int(minTemp), int(avgTemp), int(maxTemp), err
	}
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

func sendJokes(bot *tgbotapi.BotAPI) {
	text, err := getJokes()
	if err != nil {
		log.Fatal(err)
	}

	jokes1 := tgbotapi.NewMessage(reminderChatID, "Хотите анекдот?")
	_, err = bot.Send(jokes1)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(2 * time.Second)

	jokes2 := tgbotapi.NewMessage(reminderChatID, "А пофиг, слушайте")
	_, err = bot.Send(jokes2)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(2 * time.Second)

	jokes := tgbotapi.NewMessage(reminderChatID, text)
	_, err = bot.Send(jokes)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(2 * time.Second)

	jokes3 := tgbotapi.NewMessage(reminderChatID, "Ахаха")
	_, err = bot.Send(jokes3)
	if err != nil {
		log.Println(err)
	}

	lastReminderTimeMap[reminderChatID] = time.Now()
}


func sendR(bot *tgbotapi.BotAPI) {
	text, err := getJokes()
	if err != nil {
		log.Fatal(err)
	}

	jokes := tgbotapi.NewMessage(rChatID, text)
	_, err = bot.Send(jokes)
	if err != nil {
		log.Println(err)
	}

	lastReminderTimeMap[rChatID] = time.Now()
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

				patternMeet := `(?i)\b(meet|мит|миит|миток)\b`
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

				usernameWithAt := strings.ToLower("@" + bot.Self.UserName)

				rand.Seed(time.Now().UnixNano())
				switch text {
				case "да", "да)":
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
				case "нет", "нет)":
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
				case "неа", "не-а", "no", "не", "неа)", "не)":
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
				case "a", "а", "a)", "а)":
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
				case "300", "триста", "тристо", "три сотни":
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
				case "нет, тебе", "нет тебе":
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
				case "нет, ты", "нет ты":
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
				rand.Seed(time.Now().UnixNano())
				randomNumber := rand.Intn(2)
				if randomNumber == 0 {
					sendJokes(bot)
				} else {
					sendReminder(bot)
				}
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
			if (currentTime.Month() >= time.April && currentTime.Month() <= time.August && currentTime.Hour() == 7 && currentTime.Minute() == 0) ||
				(currentTime.Month() < time.April || currentTime.Month() > time.August && currentTime.Hour() == 8 && currentTime.Minute() == 0) {
				sendMorningGreetings(bot)
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

	// R send jokes loop
	go func() {
		for {
			currentTime := time.Now()
			dayOfWeek := currentTime.Weekday()
			hour := currentTime.Hour()
			if dayOfWeek >= time.Monday && dayOfWeek <= time.Friday && hour >= 8 && hour < 18 {
				lastCheckTime := lastReminderTimeMap[rChatID]
				if currentTime.Sub(lastCheckTime) >= rInterval {
					sendR(bot)
				}
			}
			time.Sleep(checkInterval)
		}
	}()

	// Keep main goroutine alive
	select {}
}
