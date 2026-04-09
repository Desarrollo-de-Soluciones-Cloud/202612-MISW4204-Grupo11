package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	appspaces"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

const errNoAuth = "no autenticado"

// endpoints para gestionar espacios académicos:
type AcademicSpaceHandler struct {
	svc *appspaces.AcademicSpaceService
}

func NewAcademicSpaceHandler(svc *appspaces.AcademicSpaceService) *AcademicSpaceHandler {
	return &AcademicSpaceHandler{svc: svc}
}

// professorIDFromContext extrae el ID del profesor autenticado del contexto de Gin.
func professorIDFromContext(c *gin.Context) (int64, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
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

// POST /spaces
func (h *AcademicSpaceHandler) Create(c *gin.Context) {
	profID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	var req createSpaceRequest
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

	space, err := h.svc.CreateSpace(c.Request.Context(), appspaces.CreateSpaceInput{
		Name:             req.Name,
		Type:             req.Type,
		AcademicPeriodID: req.AcademicPeriodID,
		ProfessorID:      profID,
		StartDate:        start,
		EndDate:          end,
		Observations:     req.Observations,
	})
	if err != nil {
		spaceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, space)
}

// GET /spaces
func (h *AcademicSpaceHandler) List(c *gin.Context) {
	profID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	spaces, err := h.svc.ListSpaces(c.Request.Context(), profID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, spaces)
}

// GET /spaces/:id
func (h *AcademicSpaceHandler) Get(c *gin.Context) {
	profID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	id, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	space, err := h.svc.GetSpace(c.Request.Context(), id, profID)
	if err != nil {
		spaceError(c, err)
		return
	}
	c.JSON(http.StatusOK, space)
}

// PATCH /spaces/:id/close
func (h *AcademicSpaceHandler) Close(c *gin.Context) {
	profID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	id, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.svc.CloseSpace(c.Request.Context(), id, profID); err != nil {
		spaceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "closed"})
}

// spaceError mapea errores de dominio a códigos HTTP.
func spaceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEspacioNoEncontrado),
		errors.Is(err, domain.ErrPeriodoNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrProfesorNoAutorizado):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrEspacioCerrado),
		errors.Is(err, domain.ErrPeriodoCerrado),
		errors.Is(err, domain.ErrTipoEspacioInvalido),
		errors.Is(err, domain.ErrFechasCierreInvalidas):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}
