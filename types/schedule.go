package types

import (
	intervalTypes "job_sender/utils/constants"
)

// Schedule holds metadata about a schedule.
type Schedule struct {
	Weekday  string `firestore:"weekday"`  // Day of week, e.g. "Monday"
	Monthday string `firestore:"monthday"` // Day of month, e.g. "1"

	Timezone string `firestore:"timezone"` // Timezone, e.g. "America/New_York"
	Time     string `firestore:"time"`     // Time of day, e.g. "09:00" in 24-hour format

	StartDate string `firestore:"start_date"` // Start date, e.g. "2021-01-01"
	EndDate   string `firestore:"end_date"`   // End date, e.g. "2021-12-31"

	IntervalType intervalTypes.IntervalTypes `firestore:"interval_type"` // The type of interval, "weeks" or "months"
	Interval     int                         `firestore:"interval"`      // The numeric interval value
}
