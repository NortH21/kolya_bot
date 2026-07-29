package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
)

type FeatureFlag string

const (
	FlagMorning  FeatureFlag = "morning"
	FlagForecast FeatureFlag = "forecast"
	FlagRates    FeatureFlag = "rates"
	FlagBenz     FeatureFlag = "benz"
	FlagAdvice   FeatureFlag = "advice"
)

var allFeatureFlags = []FeatureFlag{FlagMorning, FlagForecast, FlagRates, FlagBenz, FlagAdvice}

// Same volume kolya-bot-storage mounted in deployment.yaml, so flags survive pod restarts.
const featureFlagsFile = "/tmp/kolya_bot_flags.json"

var (
	featureFlagsMu sync.RWMutex
	featureFlags   = map[FeatureFlag]bool{
		FlagMorning:  true,
		FlagForecast: true,
		FlagRates:    true,
		FlagBenz:     true,
		FlagAdvice:   true,
	}
)

func loadFeatureFlags() {
	data, err := os.ReadFile(featureFlagsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("Ошибка чтения фича-флагов:", err)
		}
		return
	}

	var stored map[FeatureFlag]bool
	if err := json.Unmarshal(data, &stored); err != nil {
		log.Println("Ошибка разбора фича-флагов:", err)
		return
	}

	featureFlagsMu.Lock()
	defer featureFlagsMu.Unlock()
	for flag, enabled := range stored {
		if _, known := featureFlags[flag]; known {
			featureFlags[flag] = enabled
		}
	}
}

func saveFeatureFlags() {
	featureFlagsMu.RLock()
	data, err := json.MarshalIndent(featureFlags, "", "  ")
	featureFlagsMu.RUnlock()
	if err != nil {
		log.Println("Ошибка сохранения фича-флагов:", err)
		return
	}

	if err := os.WriteFile(featureFlagsFile, data, 0644); err != nil {
		log.Println("Ошибка записи файла фича-флагов:", err)
	}
}

func isFeatureEnabled(flag FeatureFlag) bool {
	featureFlagsMu.RLock()
	defer featureFlagsMu.RUnlock()
	return featureFlags[flag]
}

func setFeatureEnabled(flag FeatureFlag, enabled bool) bool {
	featureFlagsMu.Lock()
	if _, known := featureFlags[flag]; !known {
		featureFlagsMu.Unlock()
		return false
	}
	featureFlags[flag] = enabled
	featureFlagsMu.Unlock()

	saveFeatureFlags()
	return true
}

func parseFeatureFlag(name string) (FeatureFlag, bool) {
	for _, f := range allFeatureFlags {
		if string(f) == name {
			return f, true
		}
	}
	return "", false
}

func featureFlagsStatusText() string {
	featureFlagsMu.RLock()
	defer featureFlagsMu.RUnlock()

	var b strings.Builder
	b.WriteString("Фича-флаги утренней рассылки:\n")
	for _, flag := range allFeatureFlags {
		status := "✅ включено"
		if !featureFlags[flag] {
			status = "❌ выключено"
		}
		b.WriteString(string(flag) + ": " + status + "\n")
	}
	b.WriteString("\nУправление: /flag_on <название>, /flag_off <название>")
	return b.String()
}
