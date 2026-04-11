package domain

type assignment struct {
	ID                        int
	user_id                   int
	academic_space_id         int
	profesor_id               int
	role_in_assignment        string
	contracted_hours_per_week int
}
