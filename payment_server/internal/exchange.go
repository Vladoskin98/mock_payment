package internal

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type ExchangeRateParser interface {
	GetActualExchangeRate() (map[string]float64, error)
}

type MockParser struct{}

func (mp *MockParser) GetActualExchangeRate() (map[string]float64, error) {
	file, err := os.Open("./internal/CB_exchange_rate.xml")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "windows-1251" {
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		}
		return input, nil
	}

	type Valute struct {
		CharCode string `xml:"CharCode"`
		Value    string `xml:"VunitRate"`
	}

	type ValCurs struct {
		XMLName xml.Name `xml:"ValCurs"`
		Valutes []Valute `xml:"Valute"`
	}

	var valCurs ValCurs

	err = decoder.Decode(&valCurs)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга XML: %v", err)
	}

	result := make(map[string]float64)

	for _, valute := range valCurs.Valutes {
		value, err := strconv.ParseFloat(strings.ReplaceAll(valute.Value, ",", "."), 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга значения для %s: %v", valute.CharCode, err)
		}
		result[valute.CharCode] = value
	}

	return result, nil
}
