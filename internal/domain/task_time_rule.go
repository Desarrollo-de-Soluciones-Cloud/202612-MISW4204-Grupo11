package domain

import (
	"time"
)

func limitOfTimeRecordedPerRole(user *User, week int, newTaskHours int) bool {

	/*for _, vinculation := range user.Vinculations {
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
	}*/
	return false
}

func verify7daysRule(task *Task) bool {
	if time.Since(task.TimeRegistered) >= 7*24*time.Hour && task.Status == StatusOpen {
		return true
	}
	return false
}

// Para agregar a vinculaciones RF-05.2
func verifyNumberMonitoria(user *User) bool {
	/*if len(user.Vinculations) > 3 {
		return true
	}*/

	return false
}
