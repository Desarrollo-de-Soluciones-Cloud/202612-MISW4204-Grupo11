package handlers

import (
	"errors"
	"net/http"
	"strconv"

	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *apptasks.TaskService
}

func NewTaskHandler(service *apptasks.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	var task domain.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(c.Request.Context(), &task, userID); err != nil {
		taskMutateError(c, err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetAll(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	tasks, err := h.service.ListForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) ListForProfessor(c *gin.Context) {
	professorID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	tasks, err := h.service.ListForProfessor(c.Request.Context(), professorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) AdminList(c *gin.Context) {
	tasks, err := h.service.ListAllForAdmin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	taskID := c.Param("id")

	task, err := h.service.GetByIDForUser(c.Request.Context(), taskID, userID)
	if err != nil {
		taskReadError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	taskIDStr := c.Param("id")

	var task domain.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	task.ID = taskID

	if err := h.service.Update(c.Request.Context(), &task, userID); err != nil {
		taskMutateError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tarea actualizada correctamente",
	})
}

func (h *TaskHandler) UpdateField(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	taskID := c.Param("id")

	var input apptasks.UpdateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.service.PartialUpdate(c.Request.Context(), taskID, userID, input)
	if err != nil {
		taskMutateError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	taskIDStr := c.Param("id")

	var payload struct {
		Status domain.Status `json:"status"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	task := domain.Task{
		ID:     taskID,
		Status: payload.Status,
	}

	if err := h.service.UpdateStatus(c.Request.Context(), &task, userID); err != nil {
		taskMutateError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estado de la tarea actualizado correctamente",
	})
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	taskID := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), taskID, userID); err != nil {
		taskMutateError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tarea se ha eliminado correctamente",
	})
}

func (h *TaskHandler) UploadAttachment(c *gin.Context) {
	userID, ok := professorIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errNoAuth})
		return
	}

	taskID := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	attachment, err := h.service.UploadAttachment(c.Request.Context(), taskID, userID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

func taskReadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func taskMutateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrAssignmentNotOwned),
		errors.Is(err, domain.ErrTaskForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrVinculacionNoEncontrada):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
