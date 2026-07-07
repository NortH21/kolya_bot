package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var benzAPIBaseURL = "https://www.gdebenz.ru/api/nearby"

const (
	defaultBenzLat      = 57.6
	defaultBenzLon      = 39.8
	defaultBenzRadiusKm = 5
)

type nearbyStation struct {
	OSMID          string  `json:"osm_id"`
	Brand          string  `json:"brand"`
	Name           string  `json:"name"`
	Addr           string  `json:"addr"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	DistanceKm     float64 `json:"distance_km"`
	Status         string  `json:"status"`
	Detail         string  `json:"detail"`
	FuelsNow       string  `json:"fuels_now"`
	Confirmations  int     `json:"confirmations"`
	LastAt         string  `json:"last_at"`
	ConfidenceBase float64 `json:"confidence_base"`
}

type nearbyResponse struct {
	Stations []nearbyStation `json:"stations"`
}

func fetchNearbyGasStations(lat, lon float64, radiusKm int) ([]nearbyStation, error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	params.Set("radius_km", strconv.Itoa(radiusKm))

	resp, err := http.Get(benzAPIBaseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payload nearbyResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Stations, nil
}

func formatNearbyStationMessage(station nearbyStation) string {
	fuelText := strings.TrimSpace(station.FuelsNow)
	if fuelText == "" {
		fuelText = station.Status
	}
	if fuelText == "" {
		fuelText = "данные о наличии бензина отсутствуют"
	}

	address := strings.TrimSpace(station.Addr)
	if address == "" {
		address = "адрес не указан"
	}

	lastUpdate := strings.TrimSpace(station.LastAt)
	if lastUpdate == "" {
		lastUpdate = "неизвестно"
	} else if parsed, err := time.Parse("2006-01-02 15:04:05", lastUpdate); err == nil {
		lastUpdate = parsed.Format("02.01.2006 15:04")
	}

	stationName := strings.TrimSpace(station.Brand)
	if stationName == "" {
		stationName = strings.TrimSpace(station.Name)
	}
	if stationName == "" {
		stationName = "заправка"
	}

	return fmt.Sprintf("Ближайшая заправка: %s\nАдрес: %s\nБензин: %s\nОбновлено: %s", stationName, address, fuelText, lastUpdate)
}

func sendBenzInfoForCoordinates(bot *tgbotapi.BotAPI, chatID int64, lat, lon float64, radiusKm int) {
	stations, err := fetchNearbyGasStations(lat, lon, radiusKm)
	if err != nil {
		log.Printf("Ошибка запроса к api nearby: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить информацию о заправках прямо сейчас.")
		_, _ = bot.Send(msg)
		return
	}

	if len(stations) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Рядом нет заправок с доступным бензином.")
		_, _ = bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, formatNearbyStationMessage(stations[0]))
	_, _ = bot.Send(msg)
}

func sendBenzInfo(bot *tgbotapi.BotAPI, chatID int64) {
	sendBenzInfoForCoordinates(bot, chatID, defaultBenzLat, defaultBenzLon, defaultBenzRadiusKm)
}
