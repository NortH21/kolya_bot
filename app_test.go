package main

import (
	"crypto/md5"
	"encoding/hex"
	// "encoding/json"
	// "io/ioutil"
	// "net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestShouldSendReply(t *testing.T) {
	// Сбрасываем состояние перед тестом
	lastReplyTimeMap = make(map[int64]time.Time)
	replyCountMap = make(map[int64]int)

	// Устанавливаем chatID для теста
	chatID := int64(123456789)

	// Устанавливаем текущее время
	currentTime := time.Now()

	// Устанавливаем последнее время ответа для chatID на 36 минут назад
	lastReplyTimeMap[chatID] = currentTime.Add(-36 * time.Minute)

	// Проверяем, должен ли быть отправлен ответ для данного chatID
	shouldSend := shouldSendReply(chatID)
	if !shouldSend {
		t.Error("Expected shouldSend to be true, got false")
	}

	// Устанавливаем последнее время ответа на текущее время
	lastReplyTimeMap[chatID] = currentTime

	// Проверяем, что ответ не должен быть отправлен снова в течение 10 минут
	for i := 0; i < 2; i++ {
		shouldSend = shouldSendReply(chatID)
		if !shouldSend {
			t.Error("Expected shouldSend to be true, got false")
		}
	}

	// Проверяем, что ответ не должен быть отправлен снова после 3 ответов
	replyCountMap[chatID] = maxReplies
	shouldSend = shouldSendReply(chatID)
	if shouldSend {
		t.Error("Expected shouldSend to be false after max replies, got true")
	}
}

func TestShouldSendReminder(t *testing.T) {
	lastReminderTimeMap := make(map[int64]time.Time)

	// Устанавливаем последнее время напоминания для определенного chatID
	reminderChatID := int64(987654321)
	lastReminderTimeMap[reminderChatID] = time.Now().Add(-25 * time.Hour)

	// Проверяем, должно ли быть отправлено напоминание
	shouldSend := shouldSendReminder()
	currentTime := time.Now()
	if currentTime.Hour() >= 10 && currentTime.Hour() <= 20 {
		if !shouldSend {
			t.Error("Expected shouldSend to be true, got false")
		}
	}
}

func TestGetRandomLineFromFileNo(t *testing.T) {
	filename := "./files/no.txt"
	randomLine, err := getRandomLineFromFile(filename)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if randomLine == "" {
		t.Error("Expected randomLine to be non-empty, got empty")
	}
}

func TestGetRandomLineFromFileUk(t *testing.T) {
	filename := "./files/ukrf.txt"
	randomLine, err := getRandomLineFromFile(filename)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if randomLine == "" {
		t.Error("Expected randomLine to be non-empty, got empty")
	}
}

func TestGetRandomLineFromFileFriday(t *testing.T) {
	filename := "./files/friday.txt"
	randomLine, err := getRandomLineFromFile(filename)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if randomLine == "" {
		t.Error("Expected randomLine to be non-empty, got empty")
	}
}

func TestGetRandomLineFromFileReminder(t *testing.T) {
	filename := "./files/reminder.txt"
	randomLine, err := getRandomLineFromFile(filename)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if randomLine == "" {
		t.Error("Expected randomLine to be non-empty, got empty")
	}
}

func TestGetRandomLineFromFileMorning(t *testing.T) {
	filename := "./files/morning.txt"
	randomLine, err := getRandomLineFromFile(filename)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	if randomLine == "" {
		t.Error("Expected randomLine to be non-empty, got empty")
	}
}

func TestGenerateJokesURL(t *testing.T) {
	pid := "1234"
	key := "testkey"
	uts := strconv.FormatInt(time.Now().Unix(), 10)

	query := url.Values{}
	query.Set("pid", pid)
	query.Set("method", "getRandItem")
	query.Set("uts", uts)
	query.Set("category", "4") // 4 – чёрный юмор
	query.Set("genre", "1")    // 1 – анекдоты
	hash := md5.Sum([]byte(query.Encode() + key))

	expectedURL := "http://anecdotica.ru/api?category=4&genre=1&method=getRandItem&pid=1234&uts=" + uts + "&hash=" + hex.EncodeToString(hash[:])

	actualURL := generateJokesURL(pid, key)

	if actualURL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, actualURL)
	}
}

func TestIsLastDayOfMonth(t *testing.T) {
	tests := []struct {
		date       time.Time
		isLastDay  bool
	}{
		{time.Date(2024, time.April, 30, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2024, time.April, 15, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.date.String(), func(t *testing.T) {
			if got := isLastDayOfMonth(tt.date); got != tt.isLastDay {
				t.Errorf("isLastDayOfMonth(%v) = %v, want %v", tt.date, got, tt.isLastDay)
			}
		})
	}
}


// func TestGetJokes(t *testing.T) {
// 	pid := "1234"
// 	key := "testkey"

// 	url := generateJokesURL(pid, key)

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		t.Errorf("Error while sending request: %v", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Errorf("Error reading response body: %v", err)
// 		return
// 	}

// 	// Added AnecdoteResponse struct
// 	type AnecdoteResponse struct {
// 		Result struct {
// 			Error  int
// 			ErrMsg string
// 		}
// 		Item struct {
// 			Text string
// 		}
// 	}

// 	var response AnecdoteResponse
// 	err = json.Unmarshal(body, &response)
// 	if err != nil {
// 		t.Errorf("Error unmarshaling JSON: %v", err)
// 		return
// 	}

// 	if response.Result.Error != 0 {
// 		t.Errorf("Expected no error, got: %d - %s", response.Result.Error, response.Result.ErrMsg)
// 	}

// 	if response.Item.Text == "" {
// 		t.Error("Expected joke text to be non-empty")
// 	}
// }


