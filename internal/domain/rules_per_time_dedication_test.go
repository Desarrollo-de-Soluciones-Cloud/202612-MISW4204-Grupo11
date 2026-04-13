package domain_test

import (
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

func TestCheckMaxHoursPerRole_ExceedsGraduateAssistant(t *testing.T) {
	list := []domain.Assignment{
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 23},
	}
	if !domain.CheckMaxHoursPerRole(list) {
		t.Fatal("expected violation when AG > 22h")
	}
}

func TestCheckMaxHoursPerRole_ExceedsMonitor(t *testing.T) {
	list := []domain.Assignment{
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 13},
	}
	if !domain.CheckMaxHoursPerRole(list) {
		t.Fatal("expected violation when monitor > 12h")
	}
}

func TestCheckMaxHoursPerRole_WithinLimits(t *testing.T) {
	list := []domain.Assignment{
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 10},
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 20},
	}
	if domain.CheckMaxHoursPerRole(list) {
		t.Fatal("expected no violation")
	}
}

func TestLimitClasesPerUser_MoreThanThreeMonitorias(t *testing.T) {
	list := []domain.Assignment{
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
	}
	if !domain.LimitClasesPerUser(list) {
		t.Fatal("expected violation with 4 monitorías")
	}
}

func TestLimitClasesPerUser_ThreeOrFewer(t *testing.T) {
	list := []domain.Assignment{
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 10},
	}
	if domain.LimitClasesPerUser(list) {
		t.Fatal("expected no violation with 2 monitorías")
	}
}

func TestValidar40PercentOfMonitorHours_NoAssistant(t *testing.T) {
	list := []domain.Assignment{
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 20},
	}
	if domain.Validar40PercentOfMonitorHours(list) {
		t.Fatal("sin asistente no debe violar (función retorna false)")
	}
}

func TestValidar40PercentOfMonitorHours_Within40Percent(t *testing.T) {
	// 10h AG -> máximo monitor ceil(4) = 4
	list := []domain.Assignment{
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 10},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 4},
	}
	if domain.Validar40PercentOfMonitorHours(list) {
		t.Fatal("4 <= ceil(10*0.4), no violación")
	}
}

func TestValidar40PercentOfMonitorHours_Exceeds40Percent(t *testing.T) {
	// 10h AG -> máximo monitor 4; 5h monitor viola
	list := []domain.Assignment{
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 10},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 5},
	}
	if !domain.Validar40PercentOfMonitorHours(list) {
		t.Fatal("expected violation: 5 > ceil(4)")
	}
}

func TestValidar40PercentOfMonitorHours_Example22AGNineMonitor(t *testing.T) {
	// 22h AG -> ceil(8.8)=9 monitor máx; 9h OK, 10h viola
	ok := []domain.Assignment{
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 22},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 9},
	}
	if domain.Validar40PercentOfMonitorHours(ok) {
		t.Fatal("9h monitor con 22h AG debe ser válido")
	}
	bad := []domain.Assignment{
		{RoleInAssignment: "graduate_assistant", ContractedHoursPerWeek: 22},
		{RoleInAssignment: "monitor", ContractedHoursPerWeek: 10},
	}
	if !domain.Validar40PercentOfMonitorHours(bad) {
		t.Fatal("10h monitor con 22h AG debe violar")
	}
}
