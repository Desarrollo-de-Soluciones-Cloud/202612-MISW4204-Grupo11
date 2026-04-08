/*package utility

import (
	pkgtask "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/tasks"
)

func limitOfTimeRecorded(user *pkgtask.User, week int, newTaskHours int) bool {
	total := 0

	for _, vinculation := range user.Vinculations {
		if vinculation.Role != "assistant_graduated" {
			continue
		}

		for _, task := range vinculation.Tasks {
			if task.Week == week {
				total += task.TimeInvested
			}
		}
	}

	total += newTaskHours

	if total > 22 {
		return true
	}
	return false
}
*/