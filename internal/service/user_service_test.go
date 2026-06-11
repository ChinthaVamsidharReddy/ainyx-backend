package service

import (
	"testing"
	"time"
)

func TestCalculateAge(t *testing.T) {
	today := time.Now()

	tests := []struct {
		name string
		dob  time.Time
		want int
	}{
		{
			name: "birthday already passed this year",
			// Born exactly 30 years ago today minus one day → age 30
			dob:  today.AddDate(-30, 0, -1),
			want: 30,
		},
		{
			name: "birthday is today",
			dob:  today.AddDate(-25, 0, 0),
			want: 25,
		},
		{
			name: "birthday not yet this year",
			// Born 20 years ago tomorrow → age is still 19
			dob:  today.AddDate(-20, 0, 1),
			want: 19,
		},
		{
			name: "leap day birthday (28 Feb used as safe approximation)",
			dob:  time.Date(1990, time.February, 28, 0, 0, 0, 0, time.UTC),
			want: today.Year() - 1990 - boolToInt(
				today.Month() < time.February ||
					(today.Month() == time.February && today.Day() < 28),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateAge(tc.dob)
			if got != tc.want {
				t.Errorf("CalculateAge(%v) = %d; want %d", tc.dob.Format("2006-01-02"), got, tc.want)
			}
		})
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
