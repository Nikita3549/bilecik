package notifier

import (
	"strings"
	"testing"
	"time"

	"bilecik/internal/observation"

	"github.com/shopspring/decimal"
)

func hit(fromCity, toCity string, day int) observation.ThresholdHit {
	return observation.ThresholdHit{
		FromIATA:   "MSQ",
		ToIATA:     "IST",
		FromCity:   fromCity,
		ToCity:     toCity,
		FlightDate: time.Date(2026, time.February, day, 0, 0, 0, 0, time.UTC),
		Amount:     decimal.RequireFromString("1102.37"),
		Currency:   "BYN",
		Threshold:  decimal.RequireFromString("1300.00"),
	}
}

func TestFormatThresholdAlertEmpty(t *testing.T) {
	if got := formatThresholdAlert(nil); got != nil {
		t.Errorf("formatThresholdAlert(nil) = %q, want nil", got)
	}
}

func TestFormatThresholdAlertSingleMessage(t *testing.T) {
	messages := formatThresholdAlert([]observation.ThresholdHit{
		hit("Минск", "Стамбул", 20),
		hit("", "", 21),
	})

	want := alertHeader +
		"\n✈️ Минск (MSQ) → Стамбул (IST)\n📅 20 фев 2026\n💰 1 102.37 BYN (порог: 1 300 BYN)" +
		"\n\n✈️ MSQ → IST\n📅 21 фев 2026\n💰 1 102.37 BYN (порог: 1 300 BYN)"

	if len(messages) != 1 {
		t.Fatalf("got %d messages, want 1", len(messages))
	}
	if messages[0] != want {
		t.Errorf("formatThresholdAlert =\n%q\nwant\n%q", messages[0], want)
	}
}

func TestFormatThresholdAlertSplitsLongBacklog(t *testing.T) {
	hits := make([]observation.ThresholdHit, 0, 200)
	for i := range 200 {
		hits = append(hits, hit("Минск", "Стамбул", i%28+1))
	}

	messages := formatThresholdAlert(hits)

	if len(messages) < 2 {
		t.Fatalf("got %d messages, want the backlog split across several", len(messages))
	}

	totalLines := 0
	for i, m := range messages {
		// Telegram's own cap is 4096; anything above it is rejected outright.
		if len(m) > 4096 {
			t.Errorf("message %d is %d bytes, over Telegram's 4096 limit", i, len(m))
		}
		if !strings.HasPrefix(m, alertHeader) {
			t.Errorf("message %d does not start with the header: %q", i, m)
		}
		totalLines += strings.Count(m, "\n✈️ ")
	}

	if totalLines != len(hits) {
		t.Errorf("got %d flight lines across all messages, want %d", totalLines, len(hits))
	}
}
