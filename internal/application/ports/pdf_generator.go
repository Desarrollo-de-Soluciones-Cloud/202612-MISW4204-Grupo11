package ports

import (
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type PDFReportData struct {
	UserName         string
	UserEmail        string
	Role             string
	WeekStart        time.Time
	Tasks            []domain.Task
	AISummary        string
	ContractedHours  int
	TotalHoursWorked int
}

type PDFGenerator interface {
	Generate(data PDFReportData) (filePath string, err error)
}
