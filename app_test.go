package main

import (
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