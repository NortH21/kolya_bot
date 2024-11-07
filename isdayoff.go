// isdayoff.go
package main

import (
	"log"
	"time"

	"github.com/anatoliyfedorenko/isdayoff"
)

func CheckWorkday() (*isdayoff.DayType, error) {
	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCode("ru")
	pre := false
	covid := false

	var tomorrow *isdayoff.DayType
	var err error

	maxRetries := 3
	for attempts := 0; attempts < maxRetries; attempts++ {
		tomorrow, err = dayOff.Tomorrow(isdayoff.Params{
			CountryCode: &countryCode,
			Pre:         &pre,
			Covid:       &covid,
		})
		if err == nil {
			break
		}
		log.Println("Ошибка при проверке завтрашнего дня:", err)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Println("Не удалось получить данные о завтрашнем дне после нескольких попыток:", err)
		return nil, err
	}

	return tomorrow, nil
}

