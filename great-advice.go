//great-advice.go
package main

import (
	"log"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Advice struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Sound string `json:"sound"`
}

func getGreatAdvice(typega string) (string, error) {
	resp, err := http.Get("https://fucking-great-advice.ru/api/" + typega)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Ошибка: статус ответа", resp.StatusCode)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка при чтении ответа:", err)
		return "", err
	}

	var advice Advice
	if err := json.Unmarshal(body, &advice); err != nil {
		log.Println("Ошибка при разборе JSON:", err)
		return "", err
	}

	log.Println(advice.Text)
	return advice.Text, nil
}

