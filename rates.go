// rates.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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
