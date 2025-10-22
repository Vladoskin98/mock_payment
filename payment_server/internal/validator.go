package internal

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Проверка записей, в случае успеха возвращается сумма платежа в рублях, округленная до копеек.
func ValidateRequest(exchangeRate map[string]float64, providerName, amount, date string) (float64, error) {
	// Проверка providerName
	// Корректное значение -- то, которое содержит хотя бы одну букву(ограничиваюсь кириллицей и латиницей)/цифру
	re := regexp.MustCompile(`[a-zA-Zа-яА-Я0-9]`)
	if !re.MatchString(providerName) {
		return 0, fmt.Errorf("имя провайдера указано некорректно")
	}
	// Длина - не более 100 букв
	if utf8.RuneCountInString(providerName) > 100 {
		return 0, fmt.Errorf("слишком длинное имя провайдера")
	}

	// Проверка и парсинг даты. К примеру, корректная дата для 2 января 2006 года - "02.01.2006"
	_, err := time.Parse("02.01.2006", date)
	if err != nil {
		return 0, fmt.Errorf("дата указана некорректно")
	}
	if date != time.Now().Format("02.01.2006") {
		return 0, fmt.Errorf("дата должна быть сегодняшним днём")
	}

	// Проверка и парсинг суммы платежа. Пример корректного формата - "500.15 USD", после точки не более двух цифр
	amountSlice := strings.Split(amount, " ")
	if len(amountSlice) != 2 {
		return 0, fmt.Errorf("сумма платежа указана некорректно")
	}
	// Проверка дробной части суммы платежа
	if strings.Contains(amountSlice[0], ".") && utf8.RuneCountInString(strings.Split(amountSlice[0], ".")[1]) > 2 {
		return 0, fmt.Errorf("сумма платежа указана некорректно")
	}
	// Проверка наличия конкретной валюты в map с актуальным курсом
	if _, ok := exchangeRate[amountSlice[1]]; !ok {
		return 0, fmt.Errorf("банк не проводит операцию с валютой %s", amountSlice[1])
	}
	// Проверка, не указана ли неадекватно большая сумма (пусть более 20 или 0 символов)
	if utf8.RuneCountInString(strings.Split(amountSlice[0], ".")[0]) == 0 || utf8.RuneCountInString(strings.Split(amountSlice[0], ".")[0]) > 20 {
		return 0, fmt.Errorf("сумма платежа указана некорректно")
	}

	// Проверка на лимит по сумме платежа в рублях и на отрицательность числа
	parsedAmount, err := strconv.ParseFloat(amountSlice[0], 64)
	if err != nil {
		return 0, fmt.Errorf("сумма платежа указана некорректно")
	}
	if parsedAmount < 0 {
		return 0, fmt.Errorf("сумма платежа указана некорректно")
	}
	rubAmount := parsedAmount * exchangeRate[amountSlice[1]]
	rubAmount = math.Round(rubAmount*100) / 100
	if rubAmount > 15000 {
		return 0, fmt.Errorf("лимит суммы платежа - не более 15000 рублей")
	}
	return rubAmount, nil
}
