// rates.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var ratesCurrencyCodes = []string{"USD", "EUR", "AZN", "UZS", "BYN"}

type ExchangeRate struct {
	Value    float64
	Previous float64
	Nominal  float64
}

func getExchangeRate(currencyCode string) (ExchangeRate, error) {
	resp, err := http.Get("https://www.cbr-xml-daily.ru/daily_json.js")
	if err != nil {
		return ExchangeRate{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ExchangeRate{}, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ExchangeRate{}, err
	}

	valute, ok := data["Valute"].(map[string]interface{})
	if !ok {
		return ExchangeRate{}, fmt.Errorf("поле Valute не найдено в ответе ЦБ")
	}

	currency, ok := valute[currencyCode].(map[string]interface{})
	if !ok {
		return ExchangeRate{}, fmt.Errorf("валюта %s не найдена в ответе ЦБ", currencyCode)
	}

	value, ok := currency["Value"].(float64)
	if !ok {
		return ExchangeRate{}, fmt.Errorf("у валюты %s нет поля Value", currencyCode)
	}

	nominal, ok := currency["Nominal"].(float64)
	if !ok || nominal == 0 {
		nominal = 1
	}

	rate := ExchangeRate{Value: value, Nominal: nominal}
	if previous, ok := currency["Previous"].(float64); ok {
		rate.Previous = previous
	}

	return rate, nil
}

// formatRateLine форматирует строку курса. Стрелка и изменение к предыдущему
// дню добавляются только если ЦБ прислал значение Previous — если данных нет
// (например, сразу после старта или сбой в ответе ЦБ), строка выглядит как
// раньше, без стрелки.
func formatRateLine(code string, rate ExchangeRate) string {
	var base string
	if rate.Nominal != 1 {
		base = fmt.Sprintf("Курс %s (%d) к рублю: %.2f", code, int(rate.Nominal), rate.Value)
	} else {
		base = fmt.Sprintf("Курс %s к рублю: %.2f", code, rate.Value)
	}

	if rate.Previous == 0 {
		return base + "."
	}

	diff := rate.Value - rate.Previous
	arrow := "⚪️"
	if diff > 0.005 {
		arrow = "🟢▲"
	} else if diff < -0.005 {
		arrow = "🔴▼"
	}

	return fmt.Sprintf("%s %s (%+.2f).", arrow, base, diff)
}

// buildRatesMessage собирает сообщение с курсами всех валют из
// ratesCurrencyCodes. Валюты, которые не удалось получить, пропускаются
// (с логированием ошибки), а не обрушивают всё сообщение целиком.
func buildRatesMessage() string {
	var lines []string
	for _, code := range ratesCurrencyCodes {
		rate, err := getExchangeRate(code)
		if err != nil {
			log.Println(err)
			continue
		}
		lines = append(lines, formatRateLine(code, rate))
	}
	return strings.Join(lines, "\n")
}
