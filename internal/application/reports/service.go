package reports

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type ReportService struct {
	reports     domain.ReportRepository
	assignments domain.AssignmentRepository
	tasks       ports.TaskRepository
	ai          ports.AISummarizer
	pdf         ports.PDFGenerator
}

func NewReportService(
	reports domain.ReportRepository,
	assignments domain.AssignmentRepository,
	tasks ports.TaskRepository,
	ai ports.AISummarizer,
	pdf ports.PDFGenerator,
) *ReportService {
	return &ReportService{
		reports:     reports,
		assignments: assignments,
		tasks:       tasks,
		ai:          ai,
		pdf:         pdf,
	}
}

func (s *ReportService) GenerateWeeklyReports(ctx context.Context, professorID int64, weekStart time.Time) ([]domain.Report, error) {
	if err := domain.ValidateWeekStart(weekStart); err != nil {
		return nil, err
	}

	assignmentsWithUser, err := s.assignments.FindByProfessorWithUser(ctx, professorID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo vinculaciones: %w", err)
	}

	var results []domain.Report

	for _, aw := range assignmentsWithUser {
		tasks, err := s.tasks.ListByAssignmentAndWeek(ctx, aw.ID, weekStart)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo tareas para vinculación %d: %w", aw.ID, err)
		}

		if len(tasks) == 0 {
			continue
		}

		totalHours := 0
		for _, t := range tasks {
			totalHours += t.TimeInvested
		}

		prompt := buildPrompt(aw.UserName, aw.RoleInAssignment, weekStart.Format("2006-01-02"), tasks, totalHours, aw.ContractedHoursPerWeek)

		summary, err := s.ai.Summarize(ctx, prompt)
		if err != nil {
			log.Printf("AI summary failed for %s (assignment %d): %v", aw.UserName, aw.ID, err)
			summary = buildFallbackSummary(aw.UserName, aw.RoleInAssignment, weekStart, tasks, totalHours, aw.ContractedHoursPerWeek)
		}

		filePath, err := s.pdf.Generate(ports.PDFReportData{
			UserName:         aw.UserName,
			UserEmail:        aw.UserEmail,
			Role:             aw.RoleInAssignment,
			WeekStart:        weekStart,
			Tasks:            tasks,
			AISummary:        summary,
			ContractedHours:  aw.ContractedHoursPerWeek,
			TotalHoursWorked: totalHours,
		})
		if err != nil {
			return nil, fmt.Errorf("error generando PDF para %s: %w", aw.UserName, err)
		}

		report := domain.Report{
			ProfessorID:  professorID,
			AssignmentID: aw.ID,
			UserName:     aw.UserName,
			UserEmail:    aw.UserEmail,
			Role:         aw.RoleInAssignment,
			WeekStart:    weekStart,
			FilePath:     filePath,
			AISummary:    summary,
		}

		if err := s.reports.Create(ctx, &report); err != nil {
			return nil, fmt.Errorf("error almacenando reporte para %s: %w", aw.UserName, err)
		}

		results = append(results, report)
	}

	return results, nil
}

func (s *ReportService) ProcessWeeklyReportJob(ctx context.Context, job ports.WeeklyReportJob) error {
	weekStart, err := time.Parse(time.DateOnly, job.WeekStart)
	if err != nil {
		return fmt.Errorf("week_start inválido en job %s: %w", job.RequestID, err)
	}

	_, err = s.GenerateWeeklyReports(ctx, job.ProfessorID, weekStart)
	if err != nil {
		return fmt.Errorf("error procesando job %s: %w", job.RequestID, err)
	}

	return nil
}

func (s *ReportService) GetReportFile(ctx context.Context, reportID, professorID int64) (string, error) {
	report, err := s.reports.FindByID(ctx, reportID)
	if err != nil {
		return "", err
	}
	if report.ProfessorID != professorID {
		return "", domain.ErrReporteNoAutorizado
	}
	return report.FilePath, nil
}

func (s *ReportService) ListReports(ctx context.Context, professorID int64) ([]domain.Report, error) {
	return s.reports.FindByProfessor(ctx, professorID)
}

func (s *ReportService) ListReportsByWeek(ctx context.Context, professorID int64, weekStart time.Time) ([]domain.Report, error) {
	return s.reports.FindByProfessorAndWeek(ctx, professorID, weekStart)
}

// buildFallbackSummary returns a deterministic Spanish summary built only from
// the data already in the report. Used when the AI summarizer is unavailable —
// keeps the PDF informative instead of surfacing a service-down message.
func buildFallbackSummary(userName, role string, weekStart time.Time, tasks []domain.Task, totalHours, contractedHours int) string {
	weekEnd := weekStart.AddDate(0, 0, 6)
	var abierto, enDesarrollo, finalizado int
	for _, t := range tasks {
		switch t.Status {
		case domain.StatusOpen:
			abierto++
		case domain.StatusInDevelopment:
			enDesarrollo++
		case domain.StatusFinalized:
			finalizado++
		}
	}
	pct := 0
	if contractedHours > 0 {
		pct = (totalHours * 100) / contractedHours
	}
	return fmt.Sprintf(
		"Reporte semanal de %s (%s) — semana del %s al %s. "+
			"Tareas registradas: %d (%d abierta, %d en desarrollo, %d finalizada). "+
			"Horas trabajadas: %d sobre %d contratadas (%d%%).",
		userName, role,
		weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"),
		len(tasks), abierto, enDesarrollo, finalizado,
		totalHours, contractedHours, pct,
	)
}

func buildPrompt(userName, role, weekStart string, tasks []domain.Task, totalHours, contractedHours int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"Eres un asistente de supervisión académica. Analiza el siguiente reporte de actividades semanales de %s (%s) para la semana del %s.\n\n",
		userName, role, weekStart,
	))
	sb.WriteString("Tareas reportadas:\n")
	for _, t := range tasks {
		obs := t.Observations
		if obs == "" {
			obs = "ninguna"
		}
		sb.WriteString(fmt.Sprintf("- %s [%s]: %s (%dh). Observaciones: %s\n",
			t.Title, string(t.Status), t.Description, t.TimeInvested, obs))
	}
	sb.WriteString(fmt.Sprintf("\nTotal horas: %d / Contratadas: %d\n\n", totalHours, contractedHours))
	sb.WriteString("Genera un resumen breve (3-5 oraciones) en español sobre la productividad, cumplimiento de horas y observaciones relevantes. No evalúes el desempeño del estudiante.")
	return sb.String()
}
