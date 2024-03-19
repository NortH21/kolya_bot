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
	lastReplyTimeMap := make(map[int64]time.Time)

	// Устанавливаем последнее время ответа для определенного chatID
	chatID := int64(123456789)
	lastReplyTimeMap[chatID] = time.Now().Add(-30 * time.Minute)

	// Проверяем, должен ли быть отправлен ответ для данного chatID
	shouldSend := shouldSendReply(chatID)
	if !shouldSend {
		t.Error("Expected shouldSend to be true, got false")
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


