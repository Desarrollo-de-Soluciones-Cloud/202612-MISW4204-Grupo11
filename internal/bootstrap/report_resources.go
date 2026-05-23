package bootstrap

import (
	"context"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/groq"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/messaging"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/pdf"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/postgres"
	gcsstorage "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/storage/gcs"
	localstorage "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/storage/local"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type ReportResources struct {
	Pool           *postgres.Pool
	CloseDB        func()
	CloseStore     func()
	FileStorage    ports.FileStorage
	AssignmentRepo domain.AssignmentRepository
	TaskRepo       ports.TaskRepository
	ReportService  *appreports.ReportService
	RabbitMQ       *messaging.RabbitMQ
}

func BuildReportResources(ctx context.Context, cfg config.Config) (ReportResources, error) {
	pool, closeDB, err := postgres.NewPool(ctx, cfg.DBURL)
	if err != nil {
		return ReportResources{}, fmt.Errorf("postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		closeDB()
		return ReportResources{}, fmt.Errorf("postgres ping: %w", err)
	}

	db := pool.Pgx()
	if err := postgres.RunMigrations(ctx, db); err != nil {
		closeDB()
		return ReportResources{}, fmt.Errorf("migrations: %w", err)
	}

	fileStorage, closeStore, err := initStorage(ctx, cfg)
	if err != nil {
		closeDB()
		return ReportResources{}, err
	}

	assignmentRepo := postgres.NewAssignmentRepo(pool)
	taskRepo := postgres.NewTaskRepository(db)
	reportRepo := postgres.NewReportRepo(pool)
	groqClient := groq.NewClient(cfg.GroqAPIKey, cfg.GroqModel)
	pdfGenerator := pdf.NewGenerator(fileStorage, cfg.GCSReportsPrefix)
	reportService := appreports.NewReportService(reportRepo, assignmentRepo, taskRepo, groqClient, pdfGenerator)

	rabbitmqClient, err := messaging.NewRabbitMQ(cfg.BrokerURL, cfg.BrokerExchange, cfg.BrokerQueue, cfg.BrokerRoutingKey)
	if err != nil {
		closeStore()
		closeDB()
		return ReportResources{}, fmt.Errorf("rabbitmq: %w", err)
	}

	return ReportResources{
		Pool:           pool,
		CloseDB:        closeDB,
		CloseStore:     closeStore,
		FileStorage:    fileStorage,
		AssignmentRepo: assignmentRepo,
		TaskRepo:       taskRepo,
		ReportService:  reportService,
		RabbitMQ:       rabbitmqClient,
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
