![Banner del proyecto](assets/Banner.png)

# Desarrollo de Soluciones Cloud — Grupo 11

Este repositorio es el espacio de trabajo del **Grupo 11** para el proyecto del curso **Desarrollo de Soluciones Cloud (MISW4204)**. Aquí deben alojar y desarrollar cada entrega a lo largo del semestre.

## Integrantes

| Nombre | Correo |
|--------|--------|
| German Andres Gonzalez Ortega | ga.gonzalezo1@uniandes.edu.co |
| Laura Pinzon Moreno | l.pinzonm2@uniandes.edu.co |
| Santiago Mora Félix | s.moraf@uniandes.edu.co |
| Sebastian Camilo Pineda Romero | sc.pineda@uniandes.edu.co |

---

## Backend (Go, arquitectura hexagonal)

- **`cmd/api`:** arranque del programa (config, Postgres, Gin).
- **`internal/domain`:** reglas y modelos de negocio (aquí irá el corazón del sistema).
- **`internal/application`:** lógica que orquesta el dominio (ahora solo `Readiness` para comprobar la DB).
- **`internal/adapters/inbound/http`:** rutas HTTP (Gin).
- **`internal/adapters/outbound/postgres`:** conexión real a PostgreSQL.

### Cómo correrlo localmente en tu PC

1. Levanta **solo PostgreSQL**:

   ```bash
   docker compose up postgres
   ```

2. Cuando el contenedor esté *healthy*, en **otra terminal** (desde la raíz del repo), define `JWT_SECRET` y arranca:

   ```powershell
   $env:JWT_SECRET="desarrollo-cambia-esto-por-algo-largo-y-secreto"
   go run ./cmd/api
   ```

   Si no defines `DATABASE_URL`, el programa usa por defecto `127.0.0.1:5432` con usuario/clave `app` (igual que en `docker-compose.yml`).

3. Salud: [http://localhost:8080/health](http://localhost:8080/health) y [http://localhost:8080/health/ready](http://localhost:8080/health/ready).

4. **Primer administrador (base vacía)**  
   Con `users` vacío, llama `POST /api/v1/users` **sin** token. El JSON debe incluir el rol `administrador` (y la contraseña ≥ 8 caracteres). Ejemplo:

   ```json
   {
     "name": "Admin",
     "email": "tu@correo.edu.co",
     "password": "tu-clave-segura",
     "roles": ["administrador"]
   }
   ```

   Después puedes hacer `POST /api/v1/auth/login` y usar el token para crear más usuarios o listar con `GET /api/v1/users`.

5. **Autenticación y usuarios 
   - `POST /api/v1/auth/login` — body: `email`, `password` → `token` + `user`.  
   - `POST /api/v1/users` — si ya hay usuarios: `Authorization: Bearer <token>` de un **administrador**; si no hay ningún usuario: sin token, pero `roles` debe incluir `administrador`.  
   - `GET /api/v1/users` — siempre token de administrador.

   Roles globales válidos: `administrador`, `profesor`, `monitor`, `asistente_graduado`.

El API **siempre** necesita PostgreSQL accesible; si la base no está levantada, el proceso terminará con un mensaje de error al arrancar.

### Si el login devuelve 401 «correo o contraseña incorrectos»

1. **¿Creaste antes el primer usuario?** Sin filas en `users` no hay login; haz `POST /api/v1/users` con rol `administrador` (sin token).   
2. **`JWT_SECRET`:** debe ser el mismo con el que firmaste el token al hacer login.  
3. **Migraciones:** si algo falló al aplicar SQL, revisa logs del API al arrancar.

### Comandos útiles

```bash
go mod tidy
go test ./...
go vet ./...
```

### Todo con Docker (API + Postgres)

```bash
docker compose up --build
```

Mismas rutas: `GET /health` y `GET /health/ready`.

Variables: [.env.example](.env.example).


