package service

import (
	"time"
)

type Slot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func GenerateSlots(startTime string, endTime string) []Slot {
	layout := "15:04"

	start, _ := time.Parse(layout, startTime)
	end, _ := time.Parse(layout, endTime)

	var slots []Slot

	for start.Before(end) {
		next := start.Add(30 * time.Minute)

		slots = append(slots, Slot{
			Start: start.Format(layout),
			End:   next.Format(layout),
		})

		start = next
	}

	return slots
}