package util

import "testing"

func TestDate(t *testing.T) {
	tests := []struct {
		dateStr string
		valid   bool
	}{
		{"2026-01-01", true},
		{"2026-12-31", true},
		{"2026-02-29", false}, // 2026 is not a leap year
		{"2026-13-01", false}, // invalid month
		{"2026-00-01", false}, // invalid month
		{"2026-01-32", false}, // invalid day
		{"2026-01-00", false}, // invalid day
		{"67", false},
		{"flexi", false},
	}

	for _, test := range tests {
		date, err := NewDate(test.dateStr)
		if test.valid && err != nil { // no error on valid date string input
			t.Errorf("For date string %q, expected valid date but got error: %v", test.dateStr, err)
		} else if test.valid && date == nil { // non nil date on valid date string input
			t.Errorf("For date string %q, expected valid date but got nil", test.dateStr)
		} else if test.valid && date.String() != test.dateStr { // correct date.String() output
			t.Errorf("For date string %q, expected date string %q but got %q", test.dateStr, test.dateStr, date.String())
		} else if !test.valid && err == nil { // error on invalid date string input
			t.Errorf("For date string %q, expected invalid date but got valid date: %v", test.dateStr, date)
		}
	}
}
