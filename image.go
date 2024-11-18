package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/dm1trypon/go-fusionbrain-api"
	"github.com/valyala/fasthttp"
)

type CacheItem struct {
	Image      string
	CreatedAt  time.Time
}

var (
	imageCache = make(map[string]CacheItem)
	cacheMutex sync.Mutex
	cacheTTL   = 1 * time.Minute
)

func getFromCache(prompt string) (string, bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	item, found := imageCache[prompt]
	log.Println(imageCache)
	if !found {
		return "", false
	}
	if time.Since(item.CreatedAt) > cacheTTL {
		delete(imageCache, prompt)
		return "", false
	}
	return item.Image, true
}

func addToCache(prompt string, image string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	imageCache[prompt] = CacheItem{
		Image:     image,
		CreatedAt: time.Now(),
	}
}

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

	prompt = strings.ReplaceAll(prompt, "\"", "")
	log.Println("prompt: ", prompt)

	if cachedImage, found := getFromCache(prompt); found {
		log.Println("Возвращаем изображение из кеша")
		return cachedImage, nil
	}

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

	randomID := uuid.New().String()
	fileName := fmt.Sprintf("/tmp/image_%s.png", randomID)
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

	log.Print("fileName: ", fileName)

	addToCache(prompt, fileName)

	return fileName, nil
}
