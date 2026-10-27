package values

import "strings"


type CurrencyISO string

func (s CurrencyISO) ToUpper() CurrencyISO {
	return CurrencyISO(strings.ToUpper(string(s)))
}

func (s CurrencyISO) ToLower() CurrencyISO {
	return CurrencyISO(strings.ToLower(string(s)))
}

func (s CurrencyISO) String() string {
	return string(s)
}
