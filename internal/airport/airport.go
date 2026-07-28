package airport

import "strings"

type Airport struct {
	ID         int    `gorm:"column:id" json:"id"`
	Name       string `gorm:"column:name" json:"name"`
	City       string `gorm:"column:city" json:"city"`
	Country    string `gorm:"column:country" json:"country"`
	IATACode   string `gorm:"column:iata_code" json:"iata_code"`
	ICAOCode   string `gorm:"column:icao_code" json:"icao_code"`
	Language   string `gorm:"column:language" json:"language"`
	Popularity int    `gorm:"column:popularity" json:"popularity"`
}

func (Airport) TableName() string {
	return "airports"
}

// Label is city-based, not name-based: "Международный аэропорт Дубая (DXB) ·
// ОАЭ" does not fit a Telegram button, "Дубай (DXB) · ОАЭ" does.
func (a Airport) Label() string {
	place := a.City
	if place == "" {
		place = a.Name
	}
	return truncateLabel(place + " (" + a.IATACode + ") · " + a.Country)
}

// ShortLabel disambiguates airports of one city: "Москва · Шереметьево (SVO)".
// The main airport named after the city itself stays plain: "Стамбул (IST)".
func (a Airport) ShortLabel() string {
	short := a.shortName()
	if short == "" || strings.EqualFold(short, a.City) {
		return truncateLabel(a.City + " (" + a.IATACode + ")")
	}
	return truncateLabel(a.City + " · " + short + " (" + a.IATACode + ")")
}

var namePrefixes = []string{
	"Международный аэропорт ",
	"Национальный аэропорт ",
	"Лондонский аэропорт ",
	"Аэропорт ",
}

var nameSuffixes = []string{
	" International Airport",
	" Airport",
	" International",
	" аэропорт",
}

func (a Airport) shortName() string {
	s := a.Name
	for _, p := range namePrefixes {
		if strings.HasPrefix(s, p) {
			s = strings.TrimPrefix(s, p)
			break
		}
	}
	for _, suf := range nameSuffixes {
		if strings.HasSuffix(s, suf) {
			s = strings.TrimSuffix(s, suf)
			break
		}
	}
	// "Стокгольм-Арланда" / "Beijing Capital" → "Арланда" / "Capital".
	// Only when a separator follows the city, so "Ош" never eats "Ошская".
	if n := len(a.City); a.City != "" && len(s) > n && strings.EqualFold(s[:n], a.City) {
		if rest := strings.TrimLeft(s[n:], " -–"); rest != s[n:] {
			s = rest
		}
	}
	return strings.TrimSpace(s)
}

func truncateLabel(label string) string {
	if r := []rune(label); len(r) > 64 {
		return string(r[:63]) + "…"
	}
	return label
}
