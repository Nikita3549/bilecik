package money

import (
	"strings"

	"github.com/shopspring/decimal"
)

const groupSep = " "

func Format(d decimal.Decimal) string {
	return group(d.String())
}

func FormatWhole(d decimal.Decimal) string {
	return group(d.Round(0).String())
}

func group(s string) string {
	sign := ""
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		sign, s = "-", rest
	}
	whole, frac, hasFrac := strings.Cut(s, ".")

	var b strings.Builder
	b.WriteString(sign)
	for i := range len(whole) {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteString(groupSep)
		}
		b.WriteByte(whole[i])
	}
	if hasFrac {
		b.WriteByte('.')
		b.WriteString(frac)
	}
	return b.String()
}
