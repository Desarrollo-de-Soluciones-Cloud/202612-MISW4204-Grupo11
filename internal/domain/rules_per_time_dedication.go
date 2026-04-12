package domain

import "math"

//RF-05.1 Límite de asistencia graduada y RF-05.3 Límite de horas de monitoría y RF-05.6
func CheckMaxHoursPerRole(listAssignmentUser []Assignment) bool {
	totalHoursMonitor := 0
	totalHoursAsistente := 0
	for _, assignment := range listAssignmentUser {

		if assignment.RoleInAssignment == "monitor" {
			totalHoursMonitor += assignment.ContractedHoursPerWeek
		} else if assignment.RoleInAssignment == "graduate_assistant" {
			totalHoursAsistente += assignment.ContractedHoursPerWeek
		}

	}

	if totalHoursAsistente > 22 {
		return true
	} else if totalHoursMonitor > 12 {
		return true
	}

	return false
}

//RF-05.1 Límite de monitorías simultáneas y RF-05.6
func LimitClasesPerUser(listAssignmentUser []Assignment) bool {
	countMonitorias := 0

	for _, assignment := range listAssignmentUser {
		if assignment.RoleInAssignment == "monitor" {
			countMonitorias += 1
		}
	}

	if countMonitorias > 3 {
		return true
	}
	return false

}

//RF-05.4 Regla combinada de monitor, asistente graduado y RF-05.5 Redondeo del 40% y RF-05.6
func Validar40PercentOfMonitorHours(listAssignmentUser []Assignment) bool {
	totalHoursMonitor := 0
	totalHoursAsistente := 0

	for _, assignment := range listAssignmentUser {

		if assignment.RoleInAssignment == "monitor" {
			totalHoursMonitor += assignment.ContractedHoursPerWeek
		} else if assignment.RoleInAssignment == "graduate_assistant" {
			totalHoursAsistente += assignment.ContractedHoursPerWeek
		}
	}

	if totalHoursAsistente == 0 {
		return false
	}

	if float64(totalHoursMonitor) > math.Ceil(float64(totalHoursAsistente)*0.40) {
		return true
	}

	return false
}
