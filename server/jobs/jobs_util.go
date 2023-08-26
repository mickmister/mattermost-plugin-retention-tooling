package jobs

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	Monthly   Frequency = "monthly"   // Run job monthly
	Daily     Frequency = "daily"     // Run job daily
	Monday    Frequency = "monday"    // Run job weekly on Mondays
	Tuesday   Frequency = "tuesday"   // Run job weekly on Tuesdays
	Wednesday Frequency = "wednesday" // Run job weekly on Wednesdays
	Thursday  Frequency = "thursday"  // Run job weekly on Thursdays
	Friday    Frequency = "friday"    // Run job weekly on Fridays
	Saturday  Frequency = "saturday"  // Run job weekly on Saturdays
	Sunday    Frequency = "sunday"    // Run job weekly on Sundays
)

var (
	ErrInvalidFrequency = errors.New("invalid frequency")
)

type loggerIface interface {
	// Error logs an error message, optionally structured with alternating key, value parameters.
	Error(message string, keyValuePairs ...interface{})

	// Warn logs an error message, optionally structured with alternating key, value parameters.
	Warn(message string, keyValuePairs ...interface{})

	// Info logs an error message, optionally structured with alternating key, value parameters.
	Info(message string, keyValuePairs ...interface{})

	// Debug logs an error message, optionally structured with alternating key, value parameters.
	Debug(message string, keyValuePairs ...interface{})
}

type Frequency string

func FreqFromString(s string) (Frequency, error) {
	switch strings.ToLower(s) {
	case string(Monthly):
		return Monthly, nil
	case string(Daily):
		return Daily, nil
	case string(Monday):
		return Monday, nil
	case string(Tuesday):
		return Tuesday, nil
	case string(Wednesday):
		return Wednesday, nil
	case string(Thursday):
		return Thursday, nil
	case string(Friday):
		return Friday, nil
	case string(Saturday):
		return Saturday, nil
	case string(Sunday):
		return Sunday, nil
	default:
		return "", errors.Wrapf(ErrInvalidFrequency, "'%s' is not a valid frequency", s)
	}
}

// CalcNext determines the next time based on a starting time, this frequency, and the time of day option.
func (f Frequency) CalcNext(now time.Time, timeOfDay time.Time) time.Time {

}
