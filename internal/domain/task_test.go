package domain_test

import (
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

func TestTask_BelongsToCurrentWeek_CurrentMonday(t *testing.T) {
	monday := domain.CurrentWeekStart()
	task := domain.Task{WeekStart: monday}
	if !task.BelongsToCurrentWeek() {
		t.Fatal("task with CurrentWeekStart should belong to current week")
	}
}

func TestTask_BelongsToCurrentWeek_OtherWeek(t *testing.T) {
	// Lunes de hace 2 semanas
	ref := domain.CurrentWeekStart().AddDate(0, 0, -14)
	task := domain.Task{WeekStart: ref}
	if task.BelongsToCurrentWeek() {
		t.Fatal("past week should not match current week")
	}
}

func TestTask_CanBeModified_CurrentWeekNotLate(t *testing.T) {
	task := domain.Task{
		WeekStart: domain.CurrentWeekStart(),
		IsLate:    false,
	}
	if !task.CanBeModified() {
		t.Fatal("current week non-late task should be modifiable")
	}
}

func TestTask_CanBeModified_LateNotModifiable(t *testing.T) {
	task := domain.Task{
		WeekStart: domain.CurrentWeekStart(),
		IsLate:    true,
	}
	if task.CanBeModified() {
		t.Fatal("late task in current week should not be modifiable")
	}
}

func TestTask_CanBeModified_OldWeekNotModifiable(t *testing.T) {
	task := domain.Task{
		WeekStart: domain.CurrentWeekStart().AddDate(0, 0, -7),
		IsLate:    false,
	}
	if task.CanBeModified() {
		t.Fatal("task from previous week should not be modifiable as current")
	}
}

func TestCurrentWeekStart_IsMondayUTC(t *testing.T) {
	ws := domain.CurrentWeekStart()
	if ws.Weekday() != time.Monday {
		t.Fatalf("got %v, want Monday", ws.Weekday())
	}
}
