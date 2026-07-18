package cnab

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	d := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	if got := formatDate(d); got != "20072026" {
		t.Fatalf("formatDate() = %q, want %q", got, "20072026")
	}
}

func TestFormatMonthYear(t *testing.T) {
	d := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	if got := formatMonthYear(d); got != "072026" {
		t.Fatalf("formatMonthYear() = %q, want %q", got, "072026")
	}
}

func TestValidatePaymentDate(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		date    time.Time
		opts    validateOptions
		wantErr bool
	}{
		{"zero date", time.Time{}, validateOptions{}, true},
		{"past date rejected by default", now.AddDate(0, 0, -1), validateOptions{}, true},
		{"today is allowed", now, validateOptions{}, false},
		{"future date allowed", now.AddDate(0, 0, 1), validateOptions{}, false},
		{"past date allowed when AllowPastDate", now.AddDate(0, 0, -1), validateOptions{AllowPastDate: true}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validatePaymentDate(c.date, now, c.opts)
			if (err != nil) != c.wantErr {
				t.Fatalf("validatePaymentDate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}
