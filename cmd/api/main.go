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
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/groq"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/pdf"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/postgres"
	gcsstorage "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/storage/gcs"
	localstorage "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/storage/local"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	appadmin "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	appusers "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

type appResources struct {
	server     *http.Server
	rabbitmq   *messaging.RabbitMQ
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
	defer resources.rabbitmq.Close()
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
	pool, closeDB, err := postgres.NewPool(ctx, cfg.DBURL)
	if err != nil {
		return appResources{}, fmt.Errorf("postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		closeDB()
		return appResources{}, fmt.Errorf("postgres ping: %w", err)
	}

	db := pool.Pgx()
	if err := postgres.RunMigrations(ctx, db); err != nil {
		closeDB()
		return appResources{}, fmt.Errorf("migrations: %w", err)
	}

	fileStorage, closeStore, err := initStorage(ctx, cfg)
	if err != nil {
		closeDB()
		return appResources{}, err
	}

	jwtSecret := []byte(cfg.JWTSecret)
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

	taskRepo := postgres.NewTaskRepository(db)
	taskService := apptasks.NewTaskService(taskRepo, assignmentRepo).WithFileStorage(fileStorage)
	taskHandler := handlers.NewTaskHandler(taskService)

	groqClient := groq.NewClient(cfg.GroqAPIKey, cfg.GroqModel)
	pdfGenerator := pdf.NewGenerator(fileStorage, cfg.GCSReportsPrefix)
	reportRepo := postgres.NewReportRepo(pool)
	reportService := appreports.NewReportService(reportRepo, assignmentRepo, taskRepo, groqClient, pdfGenerator)

	rabbitmqClient, err := messaging.NewRabbitMQ(cfg.BrokerURL, cfg.BrokerExchange, cfg.BrokerQueue, cfg.BrokerRoutingKey)
	if err != nil {
		closeStore()
		closeDB()
		return appResources{}, fmt.Errorf("rabbitmq: %w", err)
	}
	reportSubmitService := appreports.NewSubmitService(rabbitmqClient)
	reportHandler := handlers.NewReportHandler(reportService, reportSubmitService).WithStorage(fileStorage)
	if err := rabbitmqClient.ConsumeWeeklyReportJobs(ctx, reportService.ProcessWeeklyReportJob); err != nil {
		rabbitmqClient.Close()
		closeStore()
		closeDB()
		return appResources{}, fmt.Errorf("rabbitmq consumer: %w", err)
	}

	platformOverview := appadmin.NewPlatformOverviewService(
		userRepo,
		periodRepo,
		spaceRepo,
		assignmentRepo,
		taskRepo,
	)
	adminHandler := handlers.NewAdminHandler(platformOverview)

	readiness := &application.Readiness{DB: pool}
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
		rabbitmq:   rabbitmqClient,
		closeDB:    closeDB,
		closeStore: closeStore,
	}, nil
}

func initStorage(ctx context.Context, cfg config.Config) (ports.FileStorage, func(), error) {
	switch cfg.StorageProvider {
	case "gcs":
		if cfg.GCSBucket == "" {
			return nil, func() {}, fmt.Errorf("GCS_BUCKET is required when STORAGE_PROVIDER=gcs")
		}
		gcsClient, err := gcsstorage.NewStorage(ctx, cfg.GCSBucket)
		if err != nil {
			return nil, func() {}, fmt.Errorf("gcs storage: %w", err)
		}
		return gcsClient, func() { gcsClient.Close() }, nil
	default:
		return localstorage.NewStorage(cfg.StorageLocalDir), func() {}, nil
	}
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
