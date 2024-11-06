package main

import (
	"fmt"
	"github.com/anatoliyfedorenko/isdayoff"
)

func main() {
	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCode("ru")
	pre := false
	covid := false
	day, err := dayOff.Tomorrow(isdayoff.Params{
		CountryCode: &countryCode,
		Pre:         &pre,
		Covid:       &covid,
	})    

	if err != nil {
		fmt.Println("Ошибка при проверке завтрашнего дня:", err)
		return
	}

	if *day == isdayoff.DayTypeNonWorking {
		fmt.Println("Завтра: ", *day) // 0
	} else {
		fmt.Println("Воу: ", *day) // 0
	}
	
}