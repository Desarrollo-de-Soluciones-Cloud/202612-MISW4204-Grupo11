package handlers

import (
	"errors"
	"net/http"

	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

type AssignmentHandler struct {
	svc *appspaces.AssignmentService
}

func NewAssignmentHandler(svc *appspaces.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{svc: svc}
}

type createAssignmentRequest struct {
	UserID                 int64  `json:"user_id"                   binding:"required"`
	RoleInAssignment       string `json:"role_in_assignment"        binding:"required,oneof=monitor graduate_assistant"`
	ContractedHoursPerWeek int    `json:"contracted_hours_per_week" binding:"required,min=1"`
}

func (handler *AssignmentHandler) Create(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	spaceID, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de espacio inválido"})
		return
	}

	var req createAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assignment, err := handler.svc.CreateAssignment(c.Request.Context(), appspaces.CreateAssignmentInput{
		UserID:                 req.UserID,
		AcademicSpaceID:        spaceID,
		ProfessorID:            professorID,
		RoleInAssignment:       req.RoleInAssignment,
		ContractedHoursPerWeek: req.ContractedHoursPerWeek,
	})
	if err != nil {
		assignmentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, assignment)
}


func (handler *AssignmentHandler) ListBySpace(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	spaceID, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de espacio inválido"})
		return
	}

	assignments, err := handler.svc.ListAssignmentsBySpace(c.Request.Context(), spaceID, professorID)
	if err != nil {
		assignmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, assignments)
}


func (handler *AssignmentHandler) ListMine(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	assignments, err := handler.svc.ListAssignmentsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignments)
}


func (handler *AssignmentHandler) Get(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	assignmentID, err := parseID(c, "assignmentID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de vinculación inválido"})
		return
	}

	assignment, err := handler.svc.GetAssignment(c.Request.Context(), assignmentID, professorID)
	if err != nil {
		assignmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func assignmentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrVinculacionNoEncontrada),
		errors.Is(err, domain.ErrEspacioNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrProfesorNoAutorizado):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrUsuarioYaVinculado),
		errors.Is(err, domain.ErrRolInvalido),
		errors.Is(err, domain.ErrHorasContratadas),
		errors.Is(err, domain.ErrEspacioCerradoVinculacion):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}


func (handler *AssignmentHandler) ListByProfessor(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	assignments, err := handler.svc.ListAssignmentsByProfessor(c.Request.Context(), professorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignments)
}