package main

import (
	"bufio"
	"context"
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
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	lastReplyTimeMap    map[int64]time.Time
	lastReminderTimeMap map[int64]time.Time
	lastUpdateTime      time.Time
	updateInterval            = 3 * time.Hour
	checkInterval             = 1 * time.Minute
	reminderInterval          = 14 * time.Hour
	reminderChatID      int64 = -1002039497735
	//reminderChatID int64 = 140450662
	// testId			int64 = -1001194083056
	meetUrl = "https://meet.sipleg.ru/spd"
)

var replyCountMap = make(map[int64]int)

const maxReplies = 3
const replyInterval = 35 * time.Minute

func isLastDayOfMonth(date time.Time) bool {
	nextDay := date.AddDate(0, 0, 1)
	return nextDay.Month() != date.Month()
}

func sendLastDayOfMonth(ctx context.Context, b *bot.Bot) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: reminderChatID,
		Text:   "📅 Сегодня последний день месяца! Напоминаю, что нужно заплатить за интернет.",
	})
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
	if currentTime.Hour() >= 11 && currentTime.Hour() <= 19 {
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

func sendReminder(ctx context.Context, b *bot.Bot) {
	reminderMessage, err := getRandomLineFromFile("./files/reminder.txt")
	if err != nil {
		log.Println(err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: reminderChatID,
		Text:   reminderMessage,
	})
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
	if len(lines) == 0 {
		return "", fmt.Errorf("файл %s пуст", filename)
	}
	rand.Seed(time.Now().UnixNano())
	return lines[rand.Intn(len(lines))], nil
}

func sendFridayGreetings(ctx context.Context, b *bot.Bot) {
	fridayStr, err := getRandomLineFromFile("./files/friday.txt")
	if err != nil {
		log.Println("Ошибка при получении строки из файла:", err)
		return
	}
	sendText(ctx, b, reminderChatID, fridayStr)
}

func sendMorningGreetings(ctx context.Context, b *bot.Bot) {
	morningstr, err := getRandomLineFromFile("./files/morning.txt")
	if err != nil {
		log.Println("Ошибка при получении строки из файла:", err)
		return
	}
	sendText(ctx, b, reminderChatID, morningstr)

	fullForecast, err := Forecast()
	if err == nil {
		sendText(ctx, b, reminderChatID, fullForecast)
	}

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
	sendText(ctx, b, reminderChatID, ratesstr)

	gga, err := getGreatAdvice("random")
	if err == nil {
		messageText := "Совет дня, посоны: " + gga
		sendText(ctx, b, reminderChatID, messageText)
	}
}

func shouldSendMorningGreetings(currentTime time.Time) bool {
	isSummerTime := currentTime.Month() >= time.April && currentTime.Month() <= time.August
	isSevenAM := currentTime.Hour() == 7 && currentTime.Minute() == 0
	isEightAM := currentTime.Hour() == 8 && currentTime.Minute() == 0

	return (isSummerTime && isSevenAM) || (!isSummerTime && isEightAM)
}

func sendReply(ctx context.Context, b *bot.Bot, chatID int64, replyToMessageID int, text string) {
	if shouldSendReply(chatID) {
		time.Sleep(2 * time.Second)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
			ReplyParameters: &models.ReplyParameters{
				MessageID: replyToMessageID,
			},
		})
		if err != nil {
			log.Println(err)
		}
		lastReplyTimeMap[chatID] = time.Now()
	}
}

func sendText(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func handleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	lastUpdateTime = time.Now()
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	replyToMessageID := update.Message.ID
	text := strings.ToLower(update.Message.Text)

	me, err := b.GetMe(ctx)
	if err != nil {
		log.Println("Ошибка получения информации о боте:", err)
		return
	}
	usernameWithAt := strings.ToLower("@" + me.Username)

	rand.Seed(time.Now().UnixNano())

	// regexp patterns
	patternMeet := `(?:^|\s)(meet|мит|миит|миток|meeting|хуит|хуитинг)\p{P}*(?:$|\s)`
	reMeet := regexp.MustCompile(patternMeet)
	if reMeet.MatchString(text) {
		sendText(ctx, b, chatID, "Го, я создал "+meetUrl)
	}

	patternYvn := `(?:^|\s)(ярцев|явн)\p{P}*(?:$|\s)`
	reYvn := regexp.MustCompile(patternYvn)
	if reYvn.MatchString(text) {
		sendText(ctx, b, chatID, "Самый лучший директор!")
	}

	patternUsv := `(?:^|\s)(уваров|усв|василич)\p{P}*(?:$|\s)`
	reUsv := regexp.MustCompile(patternUsv)
	if reUsv.MatchString(text) {
		sendText(ctx, b, chatID, "Тоже самый лучший директор!")
	}

	var replies = map[string]string{
		"да":           "Пизда",
		"да)":          "Пизда",
		"да!":          "Пизда",
		"мда":          "Манда",
		"мда)":         "Манда",
		"мда!":         "Манда",
		"нет":          "Пидора ответ",
		"нет)":         "Пидора ответ",
		"нет!":         "Пидора ответ",
		"a":            "Хуй на)",
		"а":            "Хуй на)",
		"a)":           "Хуй на)",
		"а)":           "Хуй на)",
		"а!":           "Хуй на)",
		"естественно":  "Хуестественно)",
		"естественно)": "Хуестественно)",
		"естественно!": "Хуестественно)",
		"чо":           "Хуй в очо)",
		"чо?":          "Хуй в очо)",
		"чо?)":         "Хуй в очо)",
		"конечно":      "Хуечно)",
		"конечно)":     "Хуечно)",
		"конечно!":     "Хуечно)",
		"300":          "Отсоси у тракториста)))",
		"триста":       "Отсоси у тракториста)))",
		"тристо":       "Отсоси у тракториста)))",
		"три сотни":    "Отсоси у тракториста)))",
		"3 сотки":      "Отсоси у тракториста)))",
		"три сотки":    "Отсоси у тракториста)))",
		"как сам":      "Как сало килограмм",
		"как сам?":     "Как сало килограмм",
		"именно":       "Хуименно",
		"именно)":      "Хуименно",
		"именно!":      "Хуименно",
		"хуй на":       "А тебе два)",
		"ну вот":       "Хуй тебе в рот)",
		"нет, тебе":    "Нет, тебе!)",
		"нет тебе":     "Нет, тебе!)",
		"нет, тебе!":   "Нет, тебе!)",
		"нет тебе!":    "Нет, тебе!)",
		"нет, ты":      "Нет, ты!)",
		"нет ты":       "Нет, ты!)",
		"нет, ты!":     "Нет, ты!)",
		"нет ты!":      "Нет, ты!)",
		"пинг":         "Хуинг",
		"ping":         "Хуинг",
		"зштп":         "Хуинг",
		"gbyu":         "Хуинг",
		"+-":           "Ты определись нахуй",
		"±":            "Ты определись нахуй",
		"-+":           "Ты определись нахуй",
		"плюс минус":   "Ты определись нахуй",
		"а то":         "а то что нахуй?",
		"а то!":        "а то что нахуй?",
	}

	if reply, ok := replies[text]; ok {
		sendReply(ctx, b, chatID, replyToMessageID, reply)
		return
	}

	switch text {
	case "/get_id", "/get_id" + usernameWithAt:
		chatIDStr := strconv.FormatInt(chatID, 10)
		log.Println("chatIDStr: " + chatIDStr)
		sendReply(ctx, b, chatID, replyToMessageID, chatIDStr)
	case "неа", "не-а", "no", "не", "неа)", "не)", "отнюдь":
		nostr, err := getRandomLineFromFile("./files/no.txt")
		if err != nil {
			log.Println(err)
		}
		sendReply(ctx, b, chatID, replyToMessageID, nostr)
	case "норм", "у меня норм", "у меня нормально", "вроде норм":
		if update.Message.Chat.Username == "Ramil4ik" {
			phrases := []string{
				"Вау, у тебя-то всё норм? Надо же, а мы тут в глуши страдаем!",
				"О, это очень помогло!",
				"Спасибо, значит ты особенный!",
				"Спасибо, Кэп!",
				"Спасибо, что поделился! Теперь я спокоен.",
			}
			rand.Seed(time.Now().UnixNano())
			randomIndex := rand.Intn(len(phrases))
			randomPhrase := phrases[randomIndex]
			sendReply(ctx, b, chatID, replyToMessageID, randomPhrase)
		}
	case "/forecast", "/forecast" + usernameWithAt:
		forecast, err := Forecast()
		if err != nil {
			log.Println(err)
		} else {
			sendText(ctx, b, chatID, forecast)
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
		sendText(ctx, b, chatID, ratesstr)
	case "/fucking_great_advice", "/fucking_great_advice" + usernameWithAt:
		fuckingGreatAdvice, err := getGreatAdvice("random")
		if err != nil {
			fmt.Println(err)
			return
		}
		sendText(ctx, b, chatID, fuckingGreatAdvice)
	case usernameWithAt:
		sendText(ctx, b, chatID, "Аа, ты меня звал? Я тут.")
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

	opts := []bot.Option{
		bot.WithDefaultHandler(handleMessage),
		bot.WithDebug(),
	}
	b, err := bot.New(os.Getenv("TELEGRAM_APITOKEN"), opts...)
	if err != nil {
		log.Fatal(err)
	}

	// Reminder loop
	go func() {
		for {
			if shouldSendReminder() {
				sendReminder(context.Background(), b)
			}
			time.Sleep(checkInterval)
		}
	}()

	// Friday/morning loop
	go func() {
		for {
			currentTime := time.Now()
			if shouldSendMorningGreetings(currentTime) {
				sendMorningGreetings(context.Background(), b)
			}

			workdayInfo, err := CheckWorkday()
			if err != nil {
				continue
			}

			if workdayInfo.Today != nil && *workdayInfo.Today == isdayoff.DayTypeWorking {
				tomorrow := workdayInfo.Tomorrow
				if tomorrow != nil && *tomorrow == isdayoff.DayTypeNonWorking {
					if currentTime.Hour() == 17 && currentTime.Minute() == 0 {
						go sendFridayGreetings(context.Background(), b)
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
				sendLastDayOfMonth(context.Background(), b)
			}
			time.Sleep(checkInterval)
		}
	}()

	b.Start(context.Background())
}
