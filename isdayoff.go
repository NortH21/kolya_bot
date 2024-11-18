package main

import (
	"log"
	"sync"
	"time"

	"github.com/anatoliyfedorenko/isdayoff"
)

type WorkdayInfo struct {
	Tomorrow *isdayoff.DayType
	Today    *isdayoff.DayType
}

type Cache struct {
	data      *WorkdayInfo
	timestamp time.Time
}

var (
	workdayCache = Cache{}
	cacheDateMutex sync.Mutex
	cacheDuration = 2 * time.Hour
)

func CheckWorkday() (*WorkdayInfo, error) {
	// log.Println("Начинаю проверять Workday")
	cacheDateMutex.Lock()
	defer cacheDateMutex.Unlock()

	if workdayCache.data != nil && time.Since(workdayCache.timestamp) < cacheDuration {
		// log.Println("Workday есть в кеше")
		return workdayCache.data, nil
	}

	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCode("ru")
	pre := false
	covid := false

	var tomorrow, today *isdayoff.DayType
	var err error

	maxRetries := 3
	for attempts := 0; attempts < maxRetries; attempts++ {
		tomorrow, err = dayOff.Tomorrow(isdayoff.Params{
			CountryCode: &countryCode,
			Pre:         &pre,
			Covid:       &covid,
		})
		if err != nil {
			log.Println("Ошибка при проверке завтрашнего дня:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		today, err = dayOff.Today(isdayoff.Params{
			CountryCode: &countryCode,
			Pre:         &pre,
			Covid:       &covid,
		})
		if err == nil {
			break
		}
		log.Println("Ошибка при проверке сегодняшнего дня:", err)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Println("Не удалось получить данные о завтрашнем или сегодняшнем дне после нескольких попыток:", err)
		return nil, err
	}

	workdayCache.data = &WorkdayInfo{
		Tomorrow: tomorrow,
		Today:    today,
	}
	workdayCache.timestamp = time.Now()

	return workdayCache.data, nil
}
