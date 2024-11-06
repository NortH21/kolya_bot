// jokees.go
package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	_ "time/tzdata"
)

func generateJokesURL(pid, key string) string {
	uts := strconv.FormatInt(time.Now().Unix(), 10)
	query := url.Values{}
	query.Set("pid", pid)
	query.Set("method", "getRandItem")
	query.Set("uts", uts)
	query.Set("category", "4") // 4 – чёрный юмор
	query.Set("genre", "1")    // 1 – анекдоты

	hash := md5.Sum([]byte(query.Encode() + key))

	u := url.URL{
		Scheme:   "http",
		Host:     "anecdotica.ru",
		Path:     "/api",
		RawQuery: query.Encode() + "&hash=" + fmt.Sprintf("%x", hash),
	}

	return u.String()
}

func getJokes() (string, error) {
	type AnecdoteResponse struct {
		Result struct {
			Error  int    `json:"error"`
			ErrMsg string `json:"errMsg"`
		} `json:"result"`
		Item struct {
			Text string `json:"text"`
			Note string `json:"note"`
		} `json:"item"`
	}

	pid := os.Getenv("anecdotica_pid")
	key := os.Getenv("anecdotica_key")

	url := generateJokesURL(pid, key)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error while sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return "", err
	}

	var anecdoteResponse AnecdoteResponse
	err = json.Unmarshal(body, &anecdoteResponse)
	if err != nil {
		log.Println("Error parsing JSON:", err)
		return "", err
	}

	if anecdoteResponse.Result.Error != 0 {
		log.Printf("API error: %s", anecdoteResponse.Result.ErrMsg)
	}

	return anecdoteResponse.Item.Text, nil
}
