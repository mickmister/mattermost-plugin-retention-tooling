package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	layoutFull      = "Jan 2, 2006 3:04pm MST"
	layoutTimeOfDay = "3:04pm MST"
)

func TestFrequency_CalcNext(t *testing.T) {
	tests := []struct {
		name      string
		f         Frequency
		last      string
		dayOfWeek int
		timeOfDay string
		want      string
	}{
		{name: "monthly sundays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 0, timeOfDay: "1:00am UTC", want: "Oct 1, 2023 1:00am UTC"},
		{name: "monthly mondays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 1, timeOfDay: "1:00am UTC", want: "Oct 2, 2023 1:00am UTC"},
		{name: "monthly tuesdays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 2, timeOfDay: "1:00am UTC", want: "Sep 26, 2023 1:00am UTC"},
		{name: "monthly wednesdays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 3, timeOfDay: "1:00am UTC", want: "Sep 27, 2023 1:00am UTC"},
		{name: "monthly thursdays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 4, timeOfDay: "1:00am UTC", want: "Sep 28, 2023 1:00am UTC"},
		{name: "monthly fridays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 5, timeOfDay: "1:00am UTC", want: "Sep 29, 2023 1:00am UTC"},
		{name: "monthly saturdays", f: Monthly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 6, timeOfDay: "1:00am UTC", want: "Sep 30, 2023 1:00am UTC"},

		{name: "weekly sundays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 0, timeOfDay: "1:00am UTC", want: "Sep 10, 2023 1:00am UTC"},
		{name: "weekly mondays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 1, timeOfDay: "1:00am UTC", want: "Sep 4, 2023 1:00am UTC"},
		{name: "weekly tuesdays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 2, timeOfDay: "1:00am UTC", want: "Sep 5, 2023 1:00am UTC"},
		{name: "weekly wednesdays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 3, timeOfDay: "1:00am UTC", want: "Sep 6, 2023 1:00am UTC"},
		{name: "weekly thursdays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 4, timeOfDay: "1:00am UTC", want: "Sep 7, 2023 1:00am UTC"},
		{name: "weekly fridays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 5, timeOfDay: "1:00am UTC", want: "Sep 8, 2023 1:00am UTC"},
		{name: "weekly saturdays (monday start)", f: Weekly, last: "Aug 28, 2023 12:48am UTC", dayOfWeek: 6, timeOfDay: "1:00am UTC", want: "Sep 9, 2023 1:00am UTC"},

		{name: "weekly sundays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 0, timeOfDay: "1:00am UTC", want: "Sep 3, 2023 1:00am UTC"},
		{name: "weekly mondays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 1, timeOfDay: "1:00am UTC", want: "Sep 4, 2023 1:00am UTC"},
		{name: "weekly tuesdays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 2, timeOfDay: "1:00am UTC", want: "Sep 5, 2023 1:00am UTC"},
		{name: "weekly wednesdays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 3, timeOfDay: "1:00am UTC", want: "Sep 6, 2023 1:00am UTC"},
		{name: "weekly thursdays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 4, timeOfDay: "1:00am UTC", want: "Sep 7, 2023 1:00am UTC"},
		{name: "weekly fridays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 5, timeOfDay: "1:00am UTC", want: "Sep 8, 2023 1:00am UTC"},
		{name: "weekly saturdays (saturday start)", f: Weekly, last: "Aug 26, 2023 12:48am UTC", dayOfWeek: 6, timeOfDay: "1:00am UTC", want: "Sep 2, 2023 1:00am UTC"},

		{name: "daily sundays", f: Daily, last: "Aug 6, 2023 12:48am UTC", dayOfWeek: 0, timeOfDay: "11:30pm UTC", want: "Aug 7, 2023 11:30pm UTC"},
		{name: "daily mondays", f: Daily, last: "Aug 7, 2023 12:48am UTC", dayOfWeek: 1, timeOfDay: "11:30pm UTC", want: "Aug 8, 2023 11:30pm UTC"},
		{name: "daily tuesdays", f: Daily, last: "Aug 8, 2023 12:48am UTC", dayOfWeek: 2, timeOfDay: "11:30pm UTC", want: "Aug 9, 2023 11:30pm UTC"},
		{name: "daily wednesdays", f: Daily, last: "Aug 9, 2023 12:48am UTC", dayOfWeek: 3, timeOfDay: "11:30pm UTC", want: "Aug 10, 2023 11:30pm UTC"},
		{name: "daily thursdays", f: Daily, last: "Aug 10, 2023 12:48am UTC", dayOfWeek: 4, timeOfDay: "11:30pm UTC", want: "Aug 11, 2023 11:30pm UTC"},
		{name: "daily fridays", f: Daily, last: "Aug 11, 2023 12:48am UTC", dayOfWeek: 5, timeOfDay: "11:30pm UTC", want: "Aug 12, 2023 11:30pm UTC"},
		{name: "daily saturdays", f: Daily, last: "Sep 30, 2023 12:48am UTC", dayOfWeek: 6, timeOfDay: "11:30pm UTC", want: "Oct 1, 2023 11:30pm UTC"},

		{name: "newyear monthly fridays", f: Monthly, last: "Dec 29, 2023 12:48am UTC", dayOfWeek: 5, timeOfDay: "1:00am UTC", want: "Feb 2, 2024 1:00am UTC"},
		{name: "newyear weekly sundays (monday start)", f: Weekly, last: "Dec 25, 2023 12:48am UTC", dayOfWeek: 0, timeOfDay: "1:00am UTC", want: "Jan 7, 2024 1:00am UTC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			last, err := time.Parse(layoutFull, tt.last)
			require.NoError(t, err)
			timeOfDay, err := time.Parse(layoutTimeOfDay, tt.timeOfDay)
			require.NoError(t, err)

			got := tt.f.CalcNext(last, tt.dayOfWeek, timeOfDay)
			gotFormatted := got.Format(layoutFull)

			assert.Equal(t, tt.want, gotFormatted)
		})
	}
}
