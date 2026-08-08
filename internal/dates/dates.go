// Package dates formats calendar dates for user-facing Telegram messages.
package dates

import (
	"fmt"
	"time"
)

var monthsShort = [...]string{"янв", "фев", "мар", "апр", "мая", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"}

// FormatDay renders «20 фев», without the year.
func FormatDay(t time.Time) string {
	return fmt.Sprintf("%d %s", t.Day(), monthsShort[t.Month()-1])
}

// FormatRange renders «20 фев», «20–22 янв», «27 янв – 12 фев», «28 дек 2026 – 3 янв 2027».
func FormatRange(from, to time.Time) string {
	sameYear := from.Year() == to.Year()

	switch {
	case from.Equal(to):
		return FormatDay(to)
	case sameYear && from.Month() == to.Month():
		return fmt.Sprintf("%d–%s", from.Day(), FormatDay(to))
	case sameYear:
		return FormatDay(from) + " – " + FormatDay(to)
	default:
		return fmt.Sprintf("%s %d – %s %d", FormatDay(from), from.Year(), FormatDay(to), to.Year())
	}
}
