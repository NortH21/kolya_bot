package main

import (
	"bytes"
	"encoding/json"
	"log"
	"io/ioutil"
	"net/http"
	"strings"
)

type RequestBody struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index int    `json:"index"`
	Delta Delta  `json:"delta"`
}

type Delta struct {
	Content string `json:"content"`
}

func Chat(text string) string {
	url := "http://ddg-chat:8787/v1/chat/completions"
	requestBody := RequestBody{
		Messages: []Message{
			{Role: "user", Content: text},
		},
		Model:  "gpt-4o-mini",
		Stream: true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Ошибка при маршализации JSON:", err)
		return ""
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return ""
	}
	defer resp.Body.Close()

	log.Printf("Статус ответа: %d\n", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка при чтении тела ответа:", err)
		return ""
	}

	var result string
	lines := strings.Split(string(body), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			jsonLine := strings.TrimPrefix(line, "data: ")
			var chunk ResponseChunk
			if err := json.Unmarshal([]byte(jsonLine), &chunk); err == nil {
				for _, choice := range chunk.Choices {
					result += choice.Delta.Content
				}
			}
		}
	}

	if result == "" {
		log.Println("Ответ пустой после декодирования.")
	}

	return result
}
