package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

const errNoAuth = "no autenticado"

type AcademicSpaceHandler struct {
	svc *appspaces.AcademicSpaceService
}

func NewAcademicSpaceHandler(svc *appspaces.AcademicSpaceService) *AcademicSpaceHandler {
	return &AcademicSpaceHandler{svc: svc}
}

func professorIDFromContext(ginCtx *gin.Context) (int64, bool) {
	userIDValue, exists := ginCtx.Get("authUserID")
	if !exists {
		return 0, false
	}
	id, ok := userIDValue.(int64)
	return id, ok
}

type createSpaceRequest struct {
	Name             string `json:"name"              binding:"required"`
	Type             string `json:"type"              binding:"required,oneof=course project"`
	AcademicPeriodID int64  `json:"academic_period_id" binding:"required"`
	StartDate        string `json:"start_date"        binding:"required"` // RFC3339 date
	EndDate          string `json:"end_date"          binding:"required"`
	Observations     string `json:"observations"`
}

func (handler *AcademicSpaceHandler) Create(ginCtx *gin.Context) {
	profID, ok := professorIDFromContext(ginCtx)
	if !ok {
		ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	var req createSpaceRequest
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

	space, err := handler.svc.CreateSpace(ginCtx.Request.Context(), appspaces.CreateSpaceInput{
		Name:             req.Name,
		Type:             req.Type,
		AcademicPeriodID: req.AcademicPeriodID,
		ProfessorID:      profID,
		StartDate:        start,
		EndDate:          end,
		Observations:     req.Observations,
	})
	if err != nil {
		spaceError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, space)
}

func (handler *AcademicSpaceHandler) List(ginCtx *gin.Context) {
	profID, ok := professorIDFromContext(ginCtx)
	if !ok {
		ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	spaces, err := handler.svc.ListSpaces(ginCtx.Request.Context(), profID)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ginCtx.JSON(http.StatusOK, spaces)
}

func (handler *AcademicSpaceHandler) Get(ginCtx *gin.Context) {
	profID, ok := professorIDFromContext(ginCtx)
	if !ok {
		ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	id, err := parseID(ginCtx, "id")
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	space, err := handler.svc.GetSpace(ginCtx.Request.Context(), id, profID)
	if err != nil {
		spaceError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, space)
}

func (handler *AcademicSpaceHandler) Close(ginCtx *gin.Context) {
	profID, ok := professorIDFromContext(ginCtx)
	if !ok {
		ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	id, err := parseID(ginCtx, "id")
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := handler.svc.CloseSpace(ginCtx.Request.Context(), id, profID); err != nil {
		spaceError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"status": "closed"})
}

func spaceError(ginCtx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEspacioNoEncontrado),
		errors.Is(err, domain.ErrPeriodoNoEncontrado):
		ginCtx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrProfesorNoAutorizado):
		ginCtx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrEspacioCerrado),
		errors.Is(err, domain.ErrPeriodoCerrado),
		errors.Is(err, domain.ErrTipoEspacioInvalido),
		errors.Is(err, domain.ErrFechasCierreInvalidas):
		ginCtx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseID(ginCtx *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(ginCtx.Param(param), 10, 64)
}
