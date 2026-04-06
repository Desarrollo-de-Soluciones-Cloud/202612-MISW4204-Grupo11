package main

import (
	"context"
	"log"

	httpadapter "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/outbound/postgres"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error leyendo configuración: %v", err)
	}

	ctx := context.Background()

	pool, closeDB, err := postgres.NewPool(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf(
			`no se pudo conectar a PostgreSQL: %v

Comprueba:
  • Que PostgreSQL esté en marcha (por ejemplo: docker compose up postgres).
  • Que el contenedor esté "healthy" antes de arrancar el API.
  • Que DATABASE_URL sea correcta si la definiste tú (usuario, contraseña, host, puerto, base "app").

Desarrollo local sin definir DATABASE_URL: se usa 127.0.0.1:5432 con usuario/clave app.`,
			err,
		)
	}
	defer closeDB()

	readiness := &application.Readiness{DB: pool}
	motor := httpadapter.NuevoMotor(readiness)

	log.Printf("Servidor en marcha. Prueba: http://localhost%s/health", cfg.HTTPAddr)

	if err := motor.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("el servidor HTTP se detuvo: %v", err)
	}
}
