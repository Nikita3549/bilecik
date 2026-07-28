package airport

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
	label := place + " (" + a.IATACode + ") · " + a.Country
	if r := []rune(label); len(r) > 64 {
		label = string(r[:63]) + "…"
	}
	return label
}
