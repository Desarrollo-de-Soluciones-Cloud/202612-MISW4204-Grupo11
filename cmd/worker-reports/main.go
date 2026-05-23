package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/bootstrap"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

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

	resources, err := bootstrap.BuildReportResources(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = resources.PubSub.Close()
		resources.CloseStore()
		resources.CloseDB()
	}()

	if err := resources.PubSub.ConsumeWeeklyReportJobs(ctx, resources.ReportService.ProcessWeeklyReportJob); err != nil {
		return fmt.Errorf("pubsub consumer: %w", err)
	}

	log.Printf("report worker ready")
	<-ctx.Done()
	return nil
}

func loadConfig() (config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
