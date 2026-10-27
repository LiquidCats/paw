package values

import (
	"fmt"
	"strings"
)

type Symbol string

func (s Symbol) ToUpper() Symbol {
	return Symbol(strings.ToUpper(string(s)))
}

func (s Symbol) ToLower() Symbol {
	return Symbol(strings.ToLower(string(s)))
}

func (s Symbol) String() string {
	return string(s)
}

func NewSymbolFromString(s string) Symbol {
	return Symbol(strings.ToUpper(s))
}

func NewSymbol(from, to CurrencyISO) Symbol {
	return Symbol(strings.ToUpper(fmt.Sprintf("%s_%s", from, to)))
}
