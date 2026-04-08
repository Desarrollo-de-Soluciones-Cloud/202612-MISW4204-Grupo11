package httpadapter

import (
	"net/http"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/middleware"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

// Deps wires HTTP routes to application services and middleware.
type Deps struct {
	Readiness *application.Readiness
	JWTSecret []byte
	Auth      *handlers.Auth
	Users     *handlers.Users
}

// NewEngine builds the Gin engine with health, auth, and user routes.
func NewEngine(deps Deps) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(ginCtx *gin.Context) {
		ginCtx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/ready", func(ginCtx *gin.Context) {
		if readyErr := deps.Readiness.Check(ginCtx.Request.Context()); readyErr != nil {
			ginCtx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"error":  readyErr.Error(),
			})
			return
		}
		ginCtx.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	apiV1 := router.Group("/api/v1")
	apiV1.POST("/auth/login", deps.Auth.PostLogin)

	// First user: POST without token (body must include administrador). Later: admin JWT only.
	apiV1.POST("/users", deps.Users.Post)

	adminUsers := apiV1.Group("/users")
	adminUsers.Use(middleware.Autenticar(deps.JWTSecret))
	adminUsers.Use(middleware.ExigeRol(domain.RolAdministrador))
	adminUsers.GET("", deps.Users.GetList)

	return router
}
