package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/messaging"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/postgres"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	appadmin "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	appusers "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/bootstrap"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

type appResources struct {
	server     *http.Server
	pubsub     *messaging.PubSubClient
	closeDB    func()
	closeStore func()
}

func main() {
	log.SetFlags(log.Ltime)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	resources, err := buildResources(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if resources.pubsub != nil {
			_ = resources.pubsub.Close()
		}
	}()
	defer resources.closeStore()
	defer resources.closeDB()

	if err := runServer(ctx, cfg.HTTPAddr, resources.server); err != nil {
		return err
	}

	return nil
}

func loadConfig() (config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, fmt.Errorf("config: %w", err)
	}
	if cfg.JWTSecret == "" {
		return config.Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func buildResources(ctx context.Context, cfg config.Config) (appResources, error) {
	resources, err := bootstrap.BuildReportResources(ctx, cfg)
	if err != nil {
		return appResources{}, err
	}

	jwtSecret := []byte(cfg.JWTSecret)
	userRepo := postgres.NewUserRepository(resources.Pool.Pgx())
	loginSvc := &auth.LoginService{Users: userRepo, Secret: jwtSecret}
	adminSvc := &appusers.AdminService{Users: userRepo}

	periodRepo := postgres.NewAcademicPeriodRepo(resources.Pool)
	spaceRepo := postgres.NewAcademicSpaceRepo(resources.Pool)
	spaceSvc := appspaces.NewAcademicSpaceService(spaceRepo, periodRepo)
	spaceHandler := handlers.NewAcademicSpaceHandler(spaceSvc)
	periodSvc := appspaces.NewAcademicPeriodService(periodRepo)
	periodHandler := handlers.NewAcademicPeriodHandler(periodSvc)

	assignmentRepo := resources.AssignmentRepo
	assignmentSvc := appspaces.NewAssignmentService(assignmentRepo, spaceRepo, periodRepo, appspaces.NoOpHourRuleChecker{})
	assignmentHandler := handlers.NewAssignmentHandler(assignmentSvc)

	taskRepo := resources.TaskRepo
	taskService := apptasks.NewTaskService(taskRepo, assignmentRepo).WithFileStorage(resources.FileStorage)
	taskHandler := handlers.NewTaskHandler(taskService)

	reportService := resources.ReportService
	pubsubClient := resources.PubSub
	reportSubmitService := appreports.NewSubmitService(pubsubClient)
	reportHandler := handlers.NewReportHandler(reportService, reportSubmitService).WithStorage(resources.FileStorage)

	platformOverview := appadmin.NewPlatformOverviewService(
		userRepo,
		periodRepo,
		spaceRepo,
		assignmentRepo,
		taskRepo,
	)
	adminHandler := handlers.NewAdminHandler(platformOverview)

	readiness := &application.Readiness{DB: resources.Pool}
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

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return appResources{
		server:     server,
		pubsub:     pubsubClient,
		closeDB:    resources.CloseDB,
		closeStore: resources.CloseStore,
	}, nil
}

func runServer(ctx context.Context, httpAddr string, server *http.Server) error {
	httpErrCh := make(chan error, 1)
	go func() {
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			httpErrCh <- serveErr
		}
	}()

	log.Printf("listening %s", httpAddr)

	select {
	case serveErr := <-httpErrCh:
		return fmt.Errorf("http: %w", serveErr)
	case <-ctx.Done():
		log.Printf("shutdown signal received")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("http graceful shutdown failed: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("http close failed: %v", closeErr)
		}
	}

	return nil
}
