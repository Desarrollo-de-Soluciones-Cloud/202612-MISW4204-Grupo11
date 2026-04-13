package domain

import "time"

// WeekStartFor returns the Monday 00:00:00 UTC of the week containing t.
func WeekStartFor(t time.Time) time.Time {
	t = t.UTC()
	offset := (int(t.Weekday()) - int(time.Monday) + 7) % 7
	y, m, d := t.AddDate(0, 0, -offset).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// CurrentWeekStart returns the Monday 00:00:00 UTC of the current week.
func CurrentWeekStart() time.Time {
	return WeekStartFor(time.Now())
}

// IsCurrentWeek returns true when weekStart equals the current week's Monday.
func IsCurrentWeek(weekStart time.Time) bool {
	return weekStart.Equal(CurrentWeekStart())
}

// ValidateWeekStart returns an error if d is not a Monday.
func ValidateWeekStart(d time.Time) error {
	if d.Weekday() != time.Monday {
		return ErrSemanaInicioNoEsLunes
	}
	return nil
}
