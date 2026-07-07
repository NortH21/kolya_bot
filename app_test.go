package main

import (
	// "crypto/md5"
	// "encoding/hex"
	// "encoding/json"
	// "io/ioutil"
	"net/http"
	"net/http/httptest"

	// "net/url"
	// "strconv"
	"strings"
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

// func TestShouldSendReminder(t *testing.T) {
// 	lastReminderTimeMap := make(map[int64]time.Time)

// 	// Устанавливаем последнее время напоминания для определенного chatID
// 	reminderChatID := int64(987654321)
// 	lastReminderTimeMap[reminderChatID] = time.Now().Add(-25 * time.Hour)

// 	// Проверяем, должно ли быть отправлено напоминание
// 	shouldSend := shouldSendReminder()
// 	currentTime := time.Now()
// 	if currentTime.Hour() >= 10 && currentTime.Hour() <= 20 {
// 		if !shouldSend {
// 			t.Error("Expected shouldSend to be true, got false")
// 		}
// 	}
// }

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

func TestIsLastDayOfMonth(t *testing.T) {
	tests := []struct {
		date      time.Time
		isLastDay bool
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

func TestFormatNearbyStationMessage(t *testing.T) {
	station := nearbyStation{
		Brand:      "Роснефть",
		Name:       "Роснефть",
		Addr:       "ул. Промзона, 11",
		DistanceKm: 0.6,
		FuelsNow:   "АИ-92, АИ-95",
		LastAt:     "2026-07-07 20:23:16",
	}

	msg := formatNearbyStationMessage(station)
	if !strings.Contains(msg, "Роснефть") {
		t.Fatalf("expected message to mention station name, got %q", msg)
	}
	if !strings.Contains(msg, "АИ-92") || !strings.Contains(msg, "ул. Промзона") {
		t.Fatalf("expected message to include fuel and address, got %q", msg)
	}
	if !strings.Contains(msg, "07.07.2026") {
		t.Fatalf("expected message to include formatted last update time, got %q", msg)
	}
}

func TestFetchNearbyGasStations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("lat"); got != "57.6" {
			t.Fatalf("expected lat query param, got %q", got)
		}
		if got := r.URL.Query().Get("lon"); got != "39.8" {
			t.Fatalf("expected lon query param, got %q", got)
		}
		if got := r.URL.Query().Get("radius_km"); got != "5" {
			t.Fatalf("expected radius_km query param, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stations":[{"brand":"Роснефть","addr":"ул. Промзона, 11","fuels_now":"АИ-92","last_at":"2026-07-07 20:23:16"}]}`))
	}))
	defer server.Close()

	oldBaseURL := benzAPIBaseURL
	benzAPIBaseURL = server.URL
	defer func() {
		benzAPIBaseURL = oldBaseURL
	}()

	stations, err := fetchNearbyGasStations(57.6, 39.8, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(stations) != 1 {
		t.Fatalf("expected one station, got %d", len(stations))
	}
	if stations[0].Addr != "ул. Промзона, 11" {
		t.Fatalf("expected parsed address, got %q", stations[0].Addr)
	}
}
