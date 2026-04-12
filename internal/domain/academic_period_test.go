package domain_test

import (
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

func TestAcademicPeriod_IsOpen_Active(t *testing.T) {
	p := domain.AcademicPeriod{
		ID:        1,
		Code:      "2026-10",
		Status:    "active",
		StartDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
	}
	if !p.IsOpen() {
		t.Fatal("status active should be open")
	}
}

func TestAcademicPeriod_IsOpen_Closed(t *testing.T) {
	p := domain.AcademicPeriod{Status: "closed"}
	if p.IsOpen() {
		t.Fatal("closed period should not be open")
	}
}

func TestAcademicPeriod_IsOpen_OtherStatus(t *testing.T) {
	statuses := []string{"", "pending", "archived"}
	for _, st := range statuses {
		p := domain.AcademicPeriod{Status: st}
		if p.IsOpen() {
			t.Fatalf("status %q should not be open", st)
		}
	}
}
