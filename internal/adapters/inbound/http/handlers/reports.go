package handlers

import (
	"errors"
	"net/http"
	"time"

	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service   *appreports.ReportService
	submitter appreports.WeeklyReportSubmitter
}

func NewReportHandler(service *appreports.ReportService, submitter ...appreports.WeeklyReportSubmitter) *ReportHandler {
	var configuredSubmitter appreports.WeeklyReportSubmitter
	if len(submitter) > 0 {
		configuredSubmitter = submitter[0]
	}
	return &ReportHandler{service: service, submitter: configuredSubmitter}
}

type generateReportRequest struct {
	WeekStart string `json:"week_start" binding:"required"`
}

func (h *ReportHandler) GenerateWeekly(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	var req generateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	weekStart, err := time.Parse(time.DateOnly, req.WeekStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "week_start inválido, use YYYY-MM-DD"})
		return
	}

	if h.submitter == nil {
		reports, err := h.service.GenerateWeeklyReports(c.Request.Context(), professorID, weekStart)
		if err != nil {
			reportError(c, err)
			return
		}
		if reports == nil {
			reports = []domain.Report{}
		}
		c.JSON(http.StatusCreated, reports)
		return
	}

	job, err := h.submitter.SubmitWeeklyReportJob(c.Request.Context(), professorID, weekStart)
	if err != nil {
		reportError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"request_id": job.RequestID,
		"status":     "queued",
		"message":    "Generación de reportes encolada",
	})
}

func (h *ReportHandler) List(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	weekStartStr := c.Query("week_start")

	if weekStartStr != "" {
		weekStart, err := time.Parse(time.DateOnly, weekStartStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "week_start inválido, use YYYY-MM-DD"})
			return
		}

		reports, err := h.service.ListReportsByWeek(c.Request.Context(), professorID, weekStart)
		if err != nil {
			reportError(c, err)
			return
		}
		c.JSON(http.StatusOK, reports)
		return
	}

	reports, err := h.service.ListReports(c.Request.Context(), professorID)
	if err != nil {
		reportError(c, err)
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) Download(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	reportID, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de reporte inválido"})
		return
	}

	filePath, err := h.service.GetReportFile(c.Request.Context(), reportID, professorID)
	if err != nil {
		reportError(c, err)
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.File(filePath)
}

func reportError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrReporteNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrReporteNoAutorizado):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrSemanaInicioNoEsLunes):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
