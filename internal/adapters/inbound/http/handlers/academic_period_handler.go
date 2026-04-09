package handlers

import (
	"errors"
	"net/http"
	"time"

	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

type AcademicPeriodHandler struct {
	svc *appspaces.AcademicPeriodService
}

func NewAcademicPeriodHandler(svc *appspaces.AcademicPeriodService) *AcademicPeriodHandler {
	return &AcademicPeriodHandler{svc: svc}
}

type createPeriodRequest struct {
	Code      string `json:"code"       binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date"   binding:"required"`
}

// POST /periods
func (h *AcademicPeriodHandler) Create(c *gin.Context) {
	var req createPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start, err := time.Parse(time.DateOnly, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date inválido, use YYYY-MM-DD"})
		return
	}
	end, err := time.Parse(time.DateOnly, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date inválido, use YYYY-MM-DD"})
		return
	}

	p, err := h.svc.CreatePeriod(c.Request.Context(), appspaces.CreatePeriodInput{
		Code:      req.Code,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		periodError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

// GET /periods
func (h *AcademicPeriodHandler) List(c *gin.Context) {
	periods, err := h.svc.ListPeriods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, periods)
}

// PATCH /periods/:id/close
func (h *AcademicPeriodHandler) Close(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := h.svc.ClosePeriod(c.Request.Context(), id); err != nil {
		periodError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "closed"})
}

func periodError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrPeriodoNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrPeriodoCerrado),
		errors.Is(err, domain.ErrFechasCierreInvalidas):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}