package util

import (
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

const layout = "2006-01-02" // YYYY-MM-DD format, see https://pkg.go.dev/time#Layout

func NewDate(dateStr string) (date *Date, err error) {
	time, err := time.Parse(layout, dateStr)
	if err != nil {
		return nil, err
	}
	date = &Date{time}
	return date, nil
}

func (d *Date) String() string {
	return d.Format(layout)
}

func (d *Date) UnmarshalJSON(b []byte) error {
	if d == nil {
		return fmt.Errorf("date: UnmarshalJSON on nil receiver")
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	if str == "" {
		d.Time = time.Time{}
		return nil
	}

	newDate, err := NewDate(str)
	if err != nil {
		return err
	}

	*d = *newDate
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(layout))
}
