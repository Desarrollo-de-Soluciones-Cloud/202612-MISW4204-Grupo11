package tasks

type Status string

const (
	open           Status = "Abierto"
	in_development Status = "En desarrollo"
	finished       Status = "fianlizado"
)

type task struct {
	ID           string
	Title        string
	Description  string
	Status       Status
	Week         int
	TimeInvested int
	Observations string
}

func newTask(title string, description string, status Status, week int, timeInvested int, observations string) *task {
	return &task{
		Title:        title,
		Description:  description,
		Status:       status,
		Week:         week,
		TimeInvested: timeInvested,
		Observations: observations,
	}
}
