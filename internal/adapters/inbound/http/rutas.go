package httpadapter

import (
	"net/http"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/middleware"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Readiness   *application.Readiness
	JWTSecret   []byte
	Auth        *handlers.Auth
	Users       *handlers.Users
	TaskHandler *handlers.TaskHandler
	AcadSpaces  *handlers.AcademicSpaceHandler
	Periods     *handlers.AcademicPeriodHandler
	Assignments *handlers.AssignmentHandler
}

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
	apiV1.POST("/users", deps.Users.Post)

	adminUsers := apiV1.Group("/users")
	adminUsers.Use(middleware.Autenticar(deps.JWTSecret))
	adminUsers.Use(middleware.ExigeRol(domain.RolAdministrador))
	adminUsers.GET("", deps.Users.GetList)

	spaces := apiV1.Group("/spaces")
	spaces.Use(middleware.Autenticar(deps.JWTSecret))
	spaces.Use(middleware.ExigeRol(domain.RolProfesor))
	{
		spaces.POST("", deps.AcadSpaces.Create)
		spaces.GET("", deps.AcadSpaces.List)
		spaces.GET("/:id", deps.AcadSpaces.Get)
		spaces.PATCH("/:id/close", deps.AcadSpaces.Close)

		spaces.POST("/:id/assignments", deps.Assignments.Create)
		spaces.GET("/:id/assignments", deps.Assignments.ListBySpace)
		spaces.GET("/:id/assignments/:assignmentID", deps.Assignments.Get)

	}

	professors := apiV1.Group("/professors")
	professors.Use(middleware.Autenticar(deps.JWTSecret))
	professors.Use(middleware.ExigeRol(domain.RolProfesor))
	professors.GET("/me/assignments", deps.Assignments.ListByProfessor)
	professors.GET("/me/tasks", deps.TaskHandler.ListForProfessor)

	periods := apiV1.Group("/periods")
	periods.Use(middleware.Autenticar(deps.JWTSecret))
	periods.Use(middleware.ExigeRol(domain.RolAdministrador))
	{
		periods.POST("", deps.Periods.Create)
		periods.GET("", deps.Periods.List)
		periods.PATCH("/:id/close", deps.Periods.Close)
	}

	tasks := apiV1.Group("/tasks")
	tasks.Use(middleware.Autenticar(deps.JWTSecret))
	tasks.Use(middleware.ExigeRol(domain.RolAsistenteGraduado, domain.RolMonitor))
	{
		tasks.POST("", deps.TaskHandler.Create)
		tasks.GET("", deps.TaskHandler.GetAll)
		tasks.PATCH("/:id", deps.TaskHandler.UpdateField)
		tasks.PUT("/:id", deps.TaskHandler.Update)
		tasks.DELETE("/:id", deps.TaskHandler.Delete)

		attachmentRoutes := tasks.Group("/:id/attachments")
		{
			attachmentRoutes.POST("", deps.TaskHandler.UploadAttachment)
		}
	}
	assignmentsMine := apiV1.Group("/assignments")
	assignmentsMine.Use(middleware.Autenticar(deps.JWTSecret))
	assignmentsMine.Use(middleware.ExigeRol(domain.RolMonitor, domain.RolAsistenteGraduado))
	assignmentsMine.GET("/me", deps.Assignments.ListMyAssignments)

	adminTasks := apiV1.Group("/admin/tasks")
	adminTasks.Use(middleware.Autenticar(deps.JWTSecret))
	adminTasks.Use(middleware.ExigeRol(domain.RolAdministrador))
	adminTasks.GET("", deps.TaskHandler.AdminList)

	adminAssignments := apiV1.Group("/admin/assignments")
	adminAssignments.Use(middleware.Autenticar(deps.JWTSecret))
	adminAssignments.Use(middleware.ExigeRol(domain.RolAdministrador))
	{
		adminAssignments.PATCH("/:assignmentID", deps.Assignments.UpdateByAdmin)
	}

	return router
}
