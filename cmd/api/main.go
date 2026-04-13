package main

import (
	"context"
	"log"

	httpadapter "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/postgres"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	appadmin "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	appusers "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/ollama"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/pdf"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

func main() {
	log.SetFlags(log.Ltime)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	jwtSecret := []byte(cfg.JWTSecret)

	ctx := context.Background()

	pool, closeDB, err := postgres.NewPool(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer closeDB()

	db := pool.Pgx()
	if err := postgres.RunMigrations(ctx, db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	loginSvc := &auth.LoginService{Users: userRepo, Secret: jwtSecret}
	adminSvc := &appusers.AdminService{Users: userRepo}

	periodRepo := postgres.NewAcademicPeriodRepo(pool)
	spaceRepo := postgres.NewAcademicSpaceRepo(pool)
	spaceSvc := appspaces.NewAcademicSpaceService(spaceRepo, periodRepo)
	spaceHandler := handlers.NewAcademicSpaceHandler(spaceSvc)
	periodSvc := appspaces.NewAcademicPeriodService(periodRepo)
	periodHandler := handlers.NewAcademicPeriodHandler(periodSvc)

	assignmentRepo := postgres.NewAssignmentRepo(pool)
	assignmentSvc := appspaces.NewAssignmentService(assignmentRepo, spaceRepo, periodRepo, appspaces.NoOpHourRuleChecker{})
	assignmentHandler := handlers.NewAssignmentHandler(assignmentSvc)

	readiness := &application.Readiness{DB: pool}

	taskRepo := postgres.NewTaskRepository(db)
	taskService := apptasks.NewTaskService(taskRepo, assignmentRepo)
	taskHandler := handlers.NewTaskHandler(taskService)

	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	ollamaClient.EnsureModel(ctx)
	pdfGenerator := pdf.NewGenerator("./uploads/reports")
	reportRepo := postgres.NewReportRepo(pool)
	reportService := appreports.NewReportService(reportRepo, assignmentRepo, taskRepo, ollamaClient, pdfGenerator)
	reportHandler := handlers.NewReportHandler(reportService)

	platformOverview := appadmin.NewPlatformOverviewService(
		userRepo,
		periodRepo,
		spaceRepo,
		assignmentRepo,
		taskRepo,
	)
	adminHandler := handlers.NewAdminHandler(platformOverview)

	engine := httpadapter.NewEngine(httpadapter.Deps{
		Readiness:   readiness,
		JWTSecret:   jwtSecret,
		Auth:        &handlers.Auth{Login: loginSvc},
		Users:       &handlers.Users{Admin: adminSvc, JWTSecret: jwtSecret},
		Admin:       adminHandler,
		TaskHandler: taskHandler,
		AcadSpaces:  spaceHandler,
		Periods:     periodHandler,
		Assignments: assignmentHandler,
		Reports:     reportHandler,
	})

	log.Printf("listening %s", cfg.HTTPAddr)

	if err := engine.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("http: %v", err)
	}

}
