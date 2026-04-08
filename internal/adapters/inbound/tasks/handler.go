package tasks

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *TaskService
}

func NewTaskHandler(service *TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) Create(c *gin.Context) {
	var task task
	var u User

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(&task, &u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetAll(c *gin.Context) {
	tasks, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
