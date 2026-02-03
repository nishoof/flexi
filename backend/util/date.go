package util

import "time"

func IsValidDate(date *string) bool {
	if date == nil {
		return false
	}
	if *date == "" {
		return false
	}
	const layout = "2006-01-02" // see https://golang.org/pkg/time/#Time.Format
	_, err := time.Parse(layout, *date)
	if err != nil {
		return false
	}
	return true
}
