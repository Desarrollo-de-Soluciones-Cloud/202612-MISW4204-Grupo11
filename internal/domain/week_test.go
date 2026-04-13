package domain_test

import (
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

func TestWeekStartFor_VariousDays(t *testing.T) {
	// Week of 2026-04-06 (Monday) through 2026-04-12 (Sunday)
	expectedMonday := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input time.Time
	}{
		{"Monday", time.Date(2026, 4, 6, 10, 30, 0, 0, time.UTC)},
		{"Tuesday", time.Date(2026, 4, 7, 8, 0, 0, 0, time.UTC)},
		{"Wednesday", time.Date(2026, 4, 8, 14, 0, 0, 0, time.UTC)},
		{"Thursday", time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)},
		{"Friday", time.Date(2026, 4, 10, 23, 59, 59, 0, time.UTC)},
		{"Saturday", time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)},
		{"Sunday", time.Date(2026, 4, 12, 18, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.WeekStartFor(tt.input)
			if !got.Equal(expectedMonday) {
				t.Errorf("WeekStartFor(%v) = %v, want %v", tt.input, got, expectedMonday)
			}
		})
	}
}

func TestWeekStartFor_NextMonday(t *testing.T) {
	// Monday 2026-04-13 should return itself
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	got := domain.WeekStartFor(monday)
	expected := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("WeekStartFor(%v) = %v, want %v", monday, got, expected)
	}
}

func TestWeekStartFor_AlwaysReturnsMonday(t *testing.T) {
	got := domain.WeekStartFor(time.Now())
	if got.Weekday() != time.Monday {
		t.Errorf("WeekStartFor(now) returned %v, want Monday", got.Weekday())
	}
}

func TestValidateWeekStart_ValidMonday(t *testing.T) {
	monday := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	if err := domain.ValidateWeekStart(monday); err != nil {
		t.Errorf("expected nil for Monday, got %v", err)
	}
}

func TestValidateWeekStart_NotMonday(t *testing.T) {
	tuesday := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)
	err := domain.ValidateWeekStart(tuesday)
	if err == nil {
		t.Fatal("expected error for Tuesday, got nil")
	}
	if err != domain.ErrSemanaInicioNoEsLunes {
		t.Errorf("expected ErrSemanaInicioNoEsLunes, got %v", err)
	}
}
