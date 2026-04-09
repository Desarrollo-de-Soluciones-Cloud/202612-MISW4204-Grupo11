package main

import (
	"context"
	"log"

	httpadapter "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http"
<<<<<<< HEAD
	taskadapter "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/tasks"
=======
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
>>>>>>> 96ecb91399457f72bcd094a4516d2a3327c4164b
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/postgres"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	appusers "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
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

	readiness := &application.Readiness{DB: pool}
<<<<<<< HEAD

	taskRepo := taskadapter.NewTaskRepository()
	taskService := taskadapter.NewTaskService(taskRepo)
	taskHandler := taskadapter.NewTaskHandler(taskService)

	motor := httpadapter.NuevoMotor(readiness, taskHandler)
=======
	engine := httpadapter.NewEngine(httpadapter.Deps{
		Readiness: readiness,
		JWTSecret: jwtSecret,
		Auth:      &handlers.Auth{Login: loginSvc},
		Users:     &handlers.Users{Admin: adminSvc, JWTSecret: jwtSecret},
	})
>>>>>>> 96ecb91399457f72bcd094a4516d2a3327c4164b

	log.Printf("listening %s", cfg.HTTPAddr)

	if err := engine.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("http: %v", err)
	}

}
