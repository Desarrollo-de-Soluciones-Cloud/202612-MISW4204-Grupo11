package tasks

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
