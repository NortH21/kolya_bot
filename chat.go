package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
	envVarChatUrl := "CHAT_URL"

	chatUrl, exists := os.LookupEnv(envVarChatUrl)
	if !exists {
		chatUrl = "ddg-chat:8787"
	}

	envVarModel := "MODEL"
	model, exists := os.LookupEnv(envVarModel)
	if !exists {
		model = "gpt-4o-mini"
	}

	url := "http://" + chatUrl + "/v1/chat/completions"
	requestBody := RequestBody{
		Messages: []Message{
			{Role: "user", Content: text},
		},
		Model:  model,
		Stream: true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Chat Ошибка при маршализации JSON:", err)
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var resp *http.Response
	var retryCount = 3
	for i := 0; i < retryCount; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("Chat Ошибка при создании запроса:", err)
			return ""
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Chat Ошибка при выполнении запроса (попытка %d): %v\n", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	if resp == nil {
		log.Println("Chat Все попытки завершились неудачей.")
		return ""
	}
	defer resp.Body.Close()

	log.Printf("Chat Статус ответа: %d\n", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Chat Ошибка при чтении тела ответа:", err)
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
		log.Println("Chat Ответ пустой после декодирования.")
	}

	return result
}
