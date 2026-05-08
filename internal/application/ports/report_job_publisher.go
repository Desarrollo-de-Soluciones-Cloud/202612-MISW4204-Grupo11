package ports

import (
	"context"
	"time"
)

type WeeklyReportJob struct {
	RequestID   string    `json:"request_id"`
	ProfessorID int64     `json:"professor_id"`
	WeekStart   string    `json:"week_start"`
	RequestedAt time.Time `json:"requested_at"`
}

type ReportJobPublisher interface {
	PublishWeeklyReportJob(ctx context.Context, job WeeklyReportJob) error
}
