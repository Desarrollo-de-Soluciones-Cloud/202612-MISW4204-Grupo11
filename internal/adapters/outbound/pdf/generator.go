package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
)

type Generator struct {
	outputDir string
}

func NewGenerator(outputDir string) *Generator {
	return &Generator{outputDir: outputDir}
}

func (g *Generator) Generate(data ports.PDFReportData) (string, error) {
	if err := os.MkdirAll(g.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("pdf mkdir: %w", err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Reporte Semanal de Actividades", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Person info
	weekEnd := data.WeekStart.AddDate(0, 0, 6)
	subtitle := fmt.Sprintf("%s (%s)", data.UserName, roleLabel(data.Role))
	weekRange := fmt.Sprintf("Semana del %s al %s", data.WeekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 7, subtitle, "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 7, weekRange, "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 7, data.UserEmail, "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// AI Summary section
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 8, "Resumen", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, data.AISummary, "", "L", false)
	pdf.Ln(6)

	// Task table
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 8, "Detalle de Tareas", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Table header
	colWidths := []float64{55, 30, 20, 85}
	headers := []string{"Titulo", "Estado", "Horas", "Descripcion"}

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(245, 245, 245)
	for i, task := range data.Tasks {
		fill := i%2 == 1
		pdf.CellFormat(colWidths[0], 7, truncate(task.Title, 30), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[1], 7, string(task.Status), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[2], 7, fmt.Sprintf("%d", task.TimeInvested), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 7, truncate(task.Description, 45), "1", 0, "L", fill, 0, "")
		pdf.Ln(-1)

		if task.Observations != "" {
			pdf.SetFont("Arial", "I", 8)
			pdf.CellFormat(colWidths[0], 6, "", "", 0, "", false, 0, "")
			pdf.MultiCell(colWidths[1]+colWidths[2]+colWidths[3], 5, "Obs: "+truncate(task.Observations, 80), "", "L", false)
			pdf.SetFont("Arial", "", 9)
		}
	}

	pdf.Ln(6)

	// Hours summary
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Total horas reportadas: %d / Horas contratadas: %d",
		data.TotalHoursWorked, data.ContractedHours), "", 1, "L", false, 0, "")

	// Save
	fileName := uuid.New().String() + ".pdf"
	filePath := filepath.Join(g.outputDir, fileName)

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", fmt.Errorf("pdf write: %w", err)
	}

	return filePath, nil
}

func roleLabel(role string) string {
	switch role {
	case "monitor":
		return "Monitor"
	case "graduate_assistant":
		return "Asistente Graduado"
	default:
		return role
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
