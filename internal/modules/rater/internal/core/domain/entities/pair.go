package entities

import (
	"fmt"
	"strings"

	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

type Pair struct {
	From   values.CurrencyISO
	To     values.CurrencyISO
	Symbol values.Symbol
}

func (p *Pair) Join(glue string) string {
	return fmt.Sprint(p.From, glue, p.To)
}

func NewPairFromSymbol(s values.Symbol) Pair {
	parts := strings.Split(s.String(), "_")
	return Pair{
		From:   values.CurrencyISO(parts[0]).ToUpper(),
		To:     values.CurrencyISO(parts[1]).ToUpper(),
		Symbol: s,
	}
}
