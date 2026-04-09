package tasks

import (
	"time"
)

func limitOfTimeRecorded22(user *User, week int, newTaskHours int) bool {
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

func verify7daysRule(task *task) bool {
	if time.Since(task.TimeRegistered) >= 7*24*time.Hour && task.Status == StatusOpen {
		return true
	}
	return false
}
