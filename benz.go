package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var benzAPIBaseURL = "https://www.gdebenz.ru/api/nearby"

var reverseGeocodeAddress = lookupAddressFromCoordinates

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

type geocodeResult struct {
	DisplayName string `json:"display_name"`
}

func lookupAddressFromCoordinates(lat, lon float64) string {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=%.6f&lon=%.6f&accept-language=ru", lat, lon)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result geocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	return strings.TrimSpace(result.DisplayName)
}

func fetchNearbyGasStations(lat, lon float64, radiusKm int) ([]nearbyStation, error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	params.Set("radius_km", strconv.Itoa(radiusKm))

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest(http.MethodGet, benzAPIBaseURL+"?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Set("Referer", "https://www.gdebenz.ru/")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < 3 {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var payload nearbyResponse
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				if attempt < 3 {
					time.Sleep(500 * time.Millisecond)
					continue
				}
				return nil, err
			}
			return payload.Stations, nil
		}

		defer resp.Body.Close()
		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		if attempt < 3 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
	}

	return nil, lastErr
}

func hasFuelAvailable(station nearbyStation) bool {
	status := strings.ToLower(strings.TrimSpace(station.Status))
	if status == "yes" {
		return true
	}
	if status == "queue" {
		return true
	}
	if status == "no" {
		return false
	}
	return strings.TrimSpace(station.FuelsNow) != ""
}

func selectBestStation(stations []nearbyStation) nearbyStation {
	var best nearbyStation
	bestFound := false
	var bestFuelCandidate nearbyStation
	bestFuelFound := false

	for _, station := range stations {
		if !bestFound || station.DistanceKm < best.DistanceKm {
			best = station
			bestFound = true
		}
		if hasFuelAvailable(station) && (!bestFuelFound || station.DistanceKm < bestFuelCandidate.DistanceKm) {
			bestFuelCandidate = station
			bestFuelFound = true
		}
	}

	if bestFuelFound {
		return bestFuelCandidate
	}
	if bestFound {
		return best
	}
	return nearbyStation{}
}

func formatFuelText(station nearbyStation) string {
	fuelText := strings.TrimSpace(station.FuelsNow)
	if fuelText != "" {
		return fuelText
	}

	status := strings.ToLower(strings.TrimSpace(station.Status))
	if status == "yes" {
		fuelText = strings.TrimSpace(station.Detail)
		if fuelText != "" {
			return fuelText
		}
		return "есть"
	}
	if status == "queue" {
		fuelText = strings.TrimSpace(station.Detail)
		if fuelText != "" {
			return fmt.Sprintf("в очереди: %s", fuelText)
		}
		return "в очереди"
	}
	if status == "no" {
		return "нет данных"
	}

	if strings.TrimSpace(station.Detail) != "" {
		return station.Detail
	}
	return "нет данных"
}

func formatNearbyStationMessage(station nearbyStation) string {
	address := strings.TrimSpace(station.Addr)
	if address == "" {
		address = reverseGeocodeAddress(station.Lat, station.Lon)
	}
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

	fuelText := formatFuelText(station)
	statusEmoji := "⛽"
	statusLabel := ""
	if strings.Contains(strings.ToLower(fuelText), "очеред") || strings.Contains(strings.ToLower(fuelText), "queue") {
		statusEmoji = "🕒"
		statusLabel = " (в очереди)"
	} else if strings.Contains(strings.ToLower(fuelText), "нет") || strings.Contains(strings.ToLower(fuelText), "нет данных") {
		statusEmoji = "⚠️"
	}

	return fmt.Sprintf("Заправка: %s%s\n📍 Адрес: %s\n%s Бензин: %s\n🕒 Обновлено: %s", stationName, statusLabel, address, statusEmoji, fuelText, lastUpdate)
}

func filterStationsForList(stations []nearbyStation) []nearbyStation {
	var filtered []nearbyStation

	for _, station := range stations {
		fuelText := strings.TrimSpace(station.FuelsNow)
		detailText := strings.TrimSpace(station.Detail)
		status := strings.ToLower(strings.TrimSpace(station.Status))
		if fuelText != "" || detailText != "" || status == "yes" || status == "queue" {
			filtered = append(filtered, station)
		}
	}

	return filtered
}

func formatNearbyStationList(stations []nearbyStation) string {
	filtered := filterStationsForList(stations)
	if len(filtered) == 0 {
		return "Пока нет данных по ближайшим заправкам."
	}

	lines := []string{"Найдены заправки:"}
	for i, station := range filtered {
		if i >= 5 {
			break
		}
		stationName := strings.TrimSpace(station.Brand)
		if stationName == "" {
			stationName = strings.TrimSpace(station.Name)
		}
		if stationName == "" {
			stationName = "заправка"
		}
		address := strings.TrimSpace(station.Addr)
		if address == "" {
			address = reverseGeocodeAddress(station.Lat, station.Lon)
		}
		if address == "" {
			address = "адрес не указан"
		}
		fuelText := formatFuelText(station)
		if strings.Contains(strings.ToLower(fuelText), "очеред") || strings.Contains(strings.ToLower(fuelText), "queue") {
			fuelText = fmt.Sprintf("%s (в очереди)", fuelText)
		}
		lines = append(lines, fmt.Sprintf("%d. %s — %s — %s", i+1, stationName, address, fuelText))
	}
	return strings.Join(lines, "\n")
}

func sendBenzInfoForCoordinates(bot *tgbotapi.BotAPI, chatID int64, lat, lon float64, radiusKm int, isListRequest bool) {
	stations, err := fetchNearbyGasStations(lat, lon, radiusKm)
	if err != nil {
		log.Printf("Ошибка запроса к api nearby: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить информацию о заправках прямо сейчас.")
		_, _ = bot.Send(msg)
		return
	}

	if len(stations) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Рядом нет заправок с доступным бензином или очередью.")
		_, _ = bot.Send(msg)
		return
	}

	if isListRequest {
		msg := tgbotapi.NewMessage(chatID, formatNearbyStationList(stations))
		_, _ = bot.Send(msg)
		return
	}

	selectedStation := selectBestStation(stations)
	msg := tgbotapi.NewMessage(chatID, formatNearbyStationMessage(selectedStation))
	_, _ = bot.Send(msg)
}

func sendBenzInfo(bot *tgbotapi.BotAPI, chatID int64) {
	sendBenzInfoForCoordinates(bot, chatID, defaultBenzLat, defaultBenzLon, defaultBenzRadiusKm, true)
}
