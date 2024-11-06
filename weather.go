//  weather.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

func getTemperature(city string) (int, int, int, int, error) {
	type Forecast struct {
		Forecastday []struct {
			Day struct {
				MaxTempC float64 `json:"maxtemp_c"`
				MinTempC float64 `json:"mintemp_c"`
				AvgTempC float64 `json:"avgtemp_c"`
			} `json:"day"`
		} `json:"forecastday"`
	}

	type Current struct {
		TempC float64 `json:"temp_c"`
	}

	type WeatherData struct {
		Current  Current  `json:"current"`
		Forecast Forecast `json:"forecast"`
	}

	url := "https://api.weatherapi.com/v1/forecast.json?q=" + city + "&days=1&key=" + os.Getenv("weatherapi_key")

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	maxTemp := weatherData.Forecast.Forecastday[0].Day.MaxTempC
	minTemp := weatherData.Forecast.Forecastday[0].Day.MinTempC
	avgTemp := weatherData.Forecast.Forecastday[0].Day.AvgTempC
	curTemp := weatherData.Current.TempC

	if city == "Yaroslavl" {
		return int(curTemp) + 2, int(minTemp) + 2, int(avgTemp) + 2, int(maxTemp) + 3, err // ну бля
	} else {
		return int(curTemp), int(minTemp), int(avgTemp), int(maxTemp), err
	}
}

