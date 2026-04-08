package httpadapter

import (
	"net/http"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/tasks"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	"github.com/gin-gonic/gin"
)

// NuevoMotor crea el servidor Gin con las rutas básicas.
func NuevoMotor(readiness *application.Readiness, taskHandler *tasks.TaskHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/health/ready", func(c *gin.Context) {
		if err := readiness.Check(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	taskRoutes := r.Group("/tasks")
	{
		taskRoutes.POST("", taskHandler.Create)
		taskRoutes.GET("", taskHandler.GetAll)
		taskRoutes.PUT("/:id", taskHandler.Update)
		taskRoutes.DELETE("/:id", taskHandler.Delete)

		attachmentRoutes := taskRoutes.Group("/:id/attachments")
		{
			attachmentRoutes.POST("", taskHandler.UploadAttachment)
		}
	}

	return r
}
