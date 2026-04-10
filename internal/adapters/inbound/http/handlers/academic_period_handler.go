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

func (handler *AcademicPeriodHandler) Create(ginCtx *gin.Context) {
	var req createPeriodRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start, err := time.Parse(time.DateOnly, req.StartDate)
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "start_date inválido, use YYYY-MM-DD"})
		return
	}
	end, err := time.Parse(time.DateOnly, req.EndDate)
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "end_date inválido, use YYYY-MM-DD"})
		return
	}

	period, err := handler.svc.CreatePeriod(ginCtx.Request.Context(), appspaces.CreatePeriodInput{
		Code:      req.Code,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		periodError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, period)
}

func (handler *AcademicPeriodHandler) List(ginCtx *gin.Context) {
	periods, err := handler.svc.ListPeriods(ginCtx.Request.Context())
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ginCtx.JSON(http.StatusOK, periods)
}

func (handler *AcademicPeriodHandler) Close(ginCtx *gin.Context) {
	id, err := parseID(ginCtx, "id")
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := handler.svc.ClosePeriod(ginCtx.Request.Context(), id); err != nil {
		periodError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"status": "closed"})
}

func periodError(ginCtx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrPeriodoNoEncontrado):
		ginCtx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrPeriodoCerrado),
		errors.Is(err, domain.ErrFechasCierreInvalidas):
		ginCtx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
