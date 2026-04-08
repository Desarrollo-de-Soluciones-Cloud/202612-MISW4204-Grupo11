package httpadapter

import (
	"net/http"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	"github.com/gin-gonic/gin"
)

// NuevoMotor crea el servidor Gin con las rutas básicas.
func NuevoMotor(
	readiness *application.Readiness,
	spaceHandler *AcademicSpaceHandler,
) *gin.Engine {
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

	// API v1 
	v1 := r.Group("/api/v1")

	//  Gestión de cursos y proyectos RF-03
	spaces := v1.Group("/spaces")
	{
		spaces.POST("", spaceHandler.Create)           // RF-03.1
		spaces.GET("", spaceHandler.List)              // RF-03.4
		spaces.GET("/:id", spaceHandler.Get)           // RF-03.4, RF-03.6
		spaces.PATCH("/:id/close", spaceHandler.Close) // RF-03.5
	}

	return r
}
