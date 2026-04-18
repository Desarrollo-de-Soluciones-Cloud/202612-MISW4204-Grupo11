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
	Admin       *handlers.Admin
	TaskHandler *handlers.TaskHandler
	AcadSpaces  *handlers.AcademicSpaceHandler
	Periods     *handlers.AcademicPeriodHandler
	Assignments *handlers.AssignmentHandler
	Reports     *handlers.ReportHandler
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

	listAssignments := apiV1.Group("userAssignments")
	listAssignments.Use(middleware.Autenticar(deps.JWTSecret))
	listAssignments.Use(middleware.ExigeRol(domain.RolAsistenteGraduado, domain.RolMonitor, domain.RolProfesor))
	{
		listAssignments.GET("", deps.Assignments.ListMyAssignments)
		listAssignments.GET("/profesor", deps.Assignments.ListByProfessor)
	}

	professors := apiV1.Group("/professors")
	professors.Use(middleware.Autenticar(deps.JWTSecret))
	professors.Use(middleware.ExigeRol(domain.RolProfesor))
	professors.GET("/me/assignments", deps.Assignments.ListByProfessor)
	professors.GET("/me/tasks", deps.TaskHandler.ListForProfessor)

	reports := apiV1.Group("/reports")
	reports.Use(middleware.Autenticar(deps.JWTSecret))
	reports.Use(middleware.ExigeRol(domain.RolProfesor))
	{
		reports.POST("/weekly", deps.Reports.GenerateWeekly)
		reports.GET("", deps.Reports.List)
		reports.GET("/:id/download", deps.Reports.Download)
	}

	periodsAdmin := apiV1.Group("/periods")
	periodsAdmin.Use(middleware.Autenticar(deps.JWTSecret))
	periodsAdmin.Use(middleware.ExigeRol(domain.RolAdministrador))
	{
		periodsAdmin.POST("", deps.Periods.Create)
		periodsAdmin.PATCH("/:id/close", deps.Periods.Close)
	}

	periodsRead := apiV1.Group("/periods")
	periodsRead.Use(middleware.Autenticar(deps.JWTSecret))
	periodsRead.Use(middleware.ExigeRol(domain.RolAdministrador, domain.RolProfesor))
	periodsRead.GET("", deps.Periods.List)

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

	adminRoot := apiV1.Group("/admin")
	adminRoot.Use(middleware.Autenticar(deps.JWTSecret))
	adminRoot.Use(middleware.ExigeRol(domain.RolAdministrador))
	{
		adminRoot.GET("/overview", deps.Admin.GetOverview)
		adminRoot.GET("/spaces", deps.AcadSpaces.ListAllForAdmin)
	}

	adminAssignments := apiV1.Group("/admin/assignments")
	adminAssignments.Use(middleware.Autenticar(deps.JWTSecret))
	adminAssignments.Use(middleware.ExigeRol(domain.RolAdministrador))
	{
		adminAssignments.GET("", deps.Assignments.ListAllForAdmin)
		adminAssignments.PATCH("/:assignmentID", deps.Assignments.UpdateByAdmin)
	}

	return router
}
