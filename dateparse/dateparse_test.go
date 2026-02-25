package dateparse

import (
	"testing"
	"time"
)

func TestParseDateExprAbsoluteDate(t *testing.T) {
	now := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)

	got, err := ParseDateExpr(now, "2025-03-01")
	if err != nil {
		t.Fatalf("ParseDateExpr() error = %v", err)
	}

	want := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("ParseDateExpr() = %v, want %v", got, want)
	}
}

func TestParseDateExprRelativeExpressions(t *testing.T) {
	now := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)

	tomorrow, err := ParseDateExpr(now, "tomorrow")
	if err != nil {
		t.Fatalf("ParseDateExpr(tomorrow) error = %v", err)
	}
	wantTomorrow := now.AddDate(0, 0, 1)
	if tomorrow.Year() != wantTomorrow.Year() || tomorrow.Month() != wantTomorrow.Month() || tomorrow.Day() != wantTomorrow.Day() {
		t.Fatalf("ParseDateExpr(tomorrow) = %v, want date %v", tomorrow, wantTomorrow)
	}

	inThreeDays, err := ParseDateExpr(now, "in 3 days")
	if err != nil {
		t.Fatalf("ParseDateExpr(in 3 days) error = %v", err)
	}
	wantInThreeDays := now.AddDate(0, 0, 3)
	if inThreeDays.Year() != wantInThreeDays.Year() || inThreeDays.Month() != wantInThreeDays.Month() || inThreeDays.Day() != wantInThreeDays.Day() {
		t.Fatalf("ParseDateExpr(in 3 days) = %v, want date %v", inThreeDays, wantInThreeDays)
	}
}

func TestParseDurationExpr(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want time.Duration
	}{
		{name: "days", expr: "1d", want: 24 * time.Hour},
		{name: "weeks", expr: "1w", want: 7 * 24 * time.Hour},
		{name: "mixed", expr: "2h30m", want: 2*time.Hour + 30*time.Minute},
		{name: "weeks and mixed", expr: "1w2d3h15m", want: 9*24*time.Hour + 3*time.Hour + 15*time.Minute},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDurationExpr(tc.expr)
			if err != nil {
				t.Fatalf("ParseDurationExpr() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("ParseDurationExpr() = %v, want %v", got, tc.want)
			}
		})
	}
}
