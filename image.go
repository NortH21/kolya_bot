package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dm1trypon/go-fusionbrain-api"
	"github.com/valyala/fasthttp"
)

func genImage(prompt string, negativePrompt string) (string, error) {
	client := &fasthttp.Client{}

	apiKey := os.Getenv("fusionbrain_key")
	apiSecret := os.Getenv("fusionbrain_secret")

	if apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("не заданы переменные окружения fusionbrain_key и fusionbrain_secret")
	}

	brain := fusionbrain.NewFusionBrain(client, apiKey, apiSecret)

	reqBody := fusionbrain.RequestBody{
		Prompt:        prompt,
		NegativePrompt: negativePrompt,
		Style:        "Детальное фото",
		Width:        1024,
		Height:       1024,
	}
	modelID := 4

	ctx := context.Background()
	uuid, err := brain.TextToImage(ctx, reqBody, modelID)
	if err != nil {
		log.Println("Ошибка при создании изображения: ", err)
		return "", err
	}
	log.Println("UUID изображения: ", uuid)

	var status fusionbrain.GenerationStatus
	for {
		status, err = brain.CheckStatus(ctx, uuid)
		if err != nil {
			log.Println("Ошибка проверки статуса: ", err)
			return "", err
		}
		if status.Status == "DONE" {
			log.Println("Изображение готово")
			break
		}
		log.Println("Изображение еще не готово, проверяем через 5 секунд")
		time.Sleep(5 * time.Second)
	}

	for i, image := range status.Images {
		if image == "" {
			log.Printf("Изображение %d пустое\n", i)
			continue
		}
		return image, nil
	}
	return "", fmt.Errorf("все изображения пустые")
}

func getImage(prompt string, negativePrompt string) (string, error) {
	img, err := genImage(prompt, negativePrompt)
	if err != nil {
		return "", err
	}

	img = strings.Trim(img, "\"")

	data, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		log.Printf("Ошибка декодирования изображения: %v\n", err)
		return "", err
	}

	fileName := "image.png"
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Ошибка создания файла %s: %v\n", fileName, err)
		return "", err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		log.Printf("Ошибка записи в файл %s: %v\n", fileName, err)
		return "", err
	}

	return fileName, nil
}
