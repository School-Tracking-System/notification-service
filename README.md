# Notification Service

Microservicio responsable del **envío y registro de notificaciones** del sistema de tracking escolar. Gestiona notificaciones push (FCM) y SMS (Twilio) hacia padres, tutores y conductores, registrando cada entrega en base de datos y permitiendo reintentos automáticos sobre entregas fallidas.

---

## Responsabilidades

- Enviar notificaciones push a dispositivos móviles vía **Firebase Cloud Messaging (FCM)**.
- Enviar mensajes SMS a tutores/padres vía **Twilio**.
- Persisitir el estado de cada notificación (`pending` → `sent` | `failed`).
- Consumir eventos de NATS para disparar notificaciones automáticamente (ej. viaje iniciado, estudiante abordado).
- Reintentar notificaciones fallidas bajo demanda.

---

## Arquitectura

```
NATS JetStream (trip.started, trip.student.boarded, ...)
      │
      │  Subscriber
      ▼
┌──────────────────────┐
│ Notification Service │  Puerto 9095 (gRPC)
│    gRPC Server       │
└──────────┬───────────┘
           │  SQL
           ▼
┌──────────────────┐     ┌───────────┐     ┌──────────────┐
│   PostgreSQL     │     │  Twilio   │     │     FCM      │
│ notification_db  │     │  (SMS)    │     │   (Push)     │
└──────────────────┘     └───────────┘     └──────────────┘
```

### Estructura Interna (Hexagonal)

```
cmd/
└── api/
    ├── main.go          # Punto de entrada. Arranca fx.App
    └── module.go        # Inyección de dependencias (Uber fx)

internal/
├── core/
│   ├── domain/
│   │   ├── models.go    # Notification, tipos y estados, constructores
│   │   └── errors.go    # ErrNotificationNotFound, ErrDeliveryFailed
│   ├── ports/
│   │   ├── repositories/  # NotificationRepository
│   │   ├── services/      # NotificationService (interfaz)
│   │   ├── resources/     # PushNotifier, SMSSender, EventPublisher
│   │   └── mocks/         # Mocks generados por mockery
│   └── notification/
│       └── notification_service.go  # Lógica de negocio
└── infrastructure/
    ├── api/
    │   └── grpc/
    │       ├── handler.go   # Mapeo Proto ↔ Domain
    │       └── server.go    # Arranque del servidor gRPC
    ├── persistence/
    │   └── postgres/
    │       └── notification_repo.go  # Implementación SQL
    ├── messaging/
    │   └── nats/
    │       ├── publisher.go    # Publicación de eventos
    │       └── subscriber.go   # Consumo de eventos (trip.*, student.*)
    └── providers/
        ├── fcm/
        │   └── push.go         # FCM PushNotifier
        └── twilio/
            └── sms.go          # Twilio SMSSender

configs/
└── db/
    └── migrations/
        └── V1__create_notifications_table.sql
```

---

## Modelo de Dominio

```go
type Notification struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Type      NotificationType    // push | sms
    Channel   string              // fcm | twilio
    Title     string
    Body      string
    Data      string             // JSON payload opcional
    Status    NotificationStatus // pending | sent | failed
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

---

## Eventos NATS Consumidos

| Sujeto NATS                       | Acción disparada                                              |
|-----------------------------------|---------------------------------------------------------------|
| `trip.started`                    | Push al conductor: "Viaje iniciado"                          |
| `trip.student.boarded`            | Push/SMS al tutor: "Estudiante ha abordado"                  |
| `trip.student.exited`             | Push/SMS al tutor: "Estudiante ha bajado"                    |
| `trip.school.reception.done`      | Push/SMS al tutor: "Estudiante recibido en la escuela"       |

---

## Requisitos Previos

- **Go** `>= 1.24`
- **PostgreSQL** `>= 14`
- **NATS Server** con JetStream habilitado
- **Cuenta Twilio** *(para SMS reales; en dev, las llamadas son simuladas)*
- **Proyecto Firebase** con credenciales de cuenta de servicio *(para push real; en dev, se loguea en consola)*
- **grpcurl** *(opcional, para pruebas manuales)*

```bash
# Instalar grpcurl (macOS)
brew install grpcurl

# NATS con JetStream
docker run -d --name nats -p 4222:4222 nats:latest -js
```

---

## Variables de Entorno

Copiar la plantilla y ajustar los valores:

```bash
cp .env.template .env
```

| Variable              | Descripción                                   | Default / Requerido                                                            |
|-----------------------|-----------------------------------------------|--------------------------------------------------------------------------------|
| `SERVICE_NAME`        | Nombre del servicio (logs/tracing)            | `notification`                                                                 |
| `PORT`                | Puerto HTTP (healthcheck futuro)              | `8085`                                                                         |
| `GRPC_PORT`           | Puerto del servidor gRPC                      | `9095`                                                                         |
| `DATABASE_URL`        | Cadena de conexión a PostgreSQL               | `postgres://postgres:postgres@localhost:5432/notification_db?sslmode=disable`  |
| `NATS_URL`            | URL del servidor NATS                         | `nats://localhost:4222`                                                        |
| `TWILIO_ACCOUNT_SID`  | SID de cuenta Twilio                          | **Requerido para SMS reales**                                                  |
| `TWILIO_AUTH_TOKEN`   | Token de autenticación Twilio                 | **Requerido para SMS reales**                                                  |
| `TWILIO_FROM_PHONE`   | Número de teléfono emisor Twilio              | **Requerido para SMS reales**                                                  |
| `FCM_CREDENTIAL_PATH` | Ruta al JSON de credenciales Firebase         | Vacío = modo dev (logs en consola)                                             |
| `ENVIRONMENT`         | Entorno (`development`/`production`)          | `development`                                                                  |
| `LOG_LEVEL`           | Nivel de log                                  | `debug`                                                                        |

> **En modo desarrollo** (`ENVIRONMENT=development`), si `FCM_CREDENTIAL_PATH` está vacío, el push notifier imprime en consola en lugar de llamar a FCM. Las credenciales de Twilio siguen siendo necesarias para SMS.

---

## Iniciar en Local

```bash
# 1. Instalar dependencias
go mod download

# 2. Copiar variables de entorno
cp .env.template .env
# Editar .env con las credenciales de Twilio y FCM si se desean notificaciones reales

# 3. Crear la base de datos
createdb notification_db

# 4. Aplicar migraciones
psql notification_db < configs/db/migrations/V1__create_notifications_table.sql

# 5. Asegurarse de tener NATS corriendo en localhost:4222

# 6. Ejecutar el servicio
go run ./cmd/api/...
```

Al iniciar, verás en los logs:
```
{"level":"info","msg":"Starting gRPC server","port":"9095"}
{"level":"info","msg":"Successfully connected to PostgreSQL"}
{"level":"info","msg":"Connected to NATS","url":"nats://localhost:4222"}
{"level":"warn","msg":"FCM running in dev mode — push notifications will be logged only"}
```

---

## API gRPC — Referencia

El contrato completo está definido en `proto/notification/v1/notification.proto`.

> **Nota sobre Swagger:** Este servicio expone únicamente gRPC (no HTTP/REST). La documentación de la API es el archivo `.proto` y los ejemplos `grpcurl` a continuación. El Gateway traduce las peticiones HTTP del cliente a llamadas gRPC hacia este servicio.

### Listar servicios disponibles

```bash
grpcurl -plaintext localhost:9095 list
```

### Enviar Notificación Push

```bash
grpcurl -plaintext -d '{
  "user_id": "<uuid-del-usuario>",
  "title":   "Viaje Iniciado",
  "body":    "El autobús escolar ha iniciado la ruta.",
  "data":    "{\"trip_id\": \"<uuid-del-viaje>\"}"
}' localhost:9095 notification.v1.NotificationService/SendPush
```

### Enviar SMS

```bash
grpcurl -plaintext -d '{
  "user_id": "<uuid-del-usuario>",
  "phone":   "+573001234567",
  "body":    "Tu estudiante ha abordado el autobús."
}' localhost:9095 notification.v1.NotificationService/SendSMS
```

### Obtener una Notificación

```bash
grpcurl -plaintext -d '{
  "id": "<uuid-de-la-notificacion>"
}' localhost:9095 notification.v1.NotificationService/GetNotification
```

### Listar Notificaciones de un Usuario

```bash
grpcurl -plaintext -d '{
  "user_id": "<uuid-del-usuario>",
  "limit":   20,
  "offset":  0
}' localhost:9095 notification.v1.NotificationService/ListNotifications
```

Con filtro por estado:

```bash
grpcurl -plaintext -d '{
  "user_id": "<uuid-del-usuario>",
  "status":  "failed",
  "limit":   10
}' localhost:9095 notification.v1.NotificationService/ListNotifications
```

### Reintentar Notificaciones Fallidas

```bash
grpcurl -plaintext -d '{}' localhost:9095 notification.v1.NotificationService/RetryFailed
```

---

## Probar con grpcui (Interfaz Visual — equivalente a Swagger UI para gRPC)

`grpcui` ofrece una interfaz web interactiva para explorar y llamar métodos gRPC, similar a Swagger UI:

```bash
# Instalar grpcui (macOS)
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

# Abrir la UI en el navegador (el servicio debe estar corriendo)
grpcui -plaintext localhost:9095
```

Esto abrirá automáticamente `http://localhost:8080` con un formulario interactivo para cada método del servicio.

---

## Documentación de la API

El contrato gRPC está documentado en dos lugares:

| Formato | Ruta | Uso |
|---|---|---|
| **Proto** (fuente de verdad) | `proto/notification/v1/notification.proto` | Contrato oficial |
| **OpenAPI / Swagger** | `docs/api/swagger.yaml` | Referencia visual (convención gRPC-over-HTTP) |

Para visualizar el `swagger.yaml` en Swagger UI:
```bash
# Con Docker
docker run -p 8081:8080 \
  -e SWAGGER_JSON=/docs/swagger.yaml \
  -v $(pwd)/docs/api:/docs \
  swaggerapi/swagger-ui
# Abrir: http://localhost:8081
```

---

## Ejecutar Tests

```bash
go test ./internal/core/notification/... -v
```

---

## Tecnologías Utilizadas

| Librería                          | Propósito                                    |
|-----------------------------------|----------------------------------------------|
| `google.golang.org/grpc`          | Servidor gRPC                                |
| `google.golang.org/protobuf`      | Serialización Protocol Buffers               |
| `go.uber.org/fx`                  | Inyección de dependencias y ciclo de vida    |
| `go.uber.org/zap`                 | Logging estructurado                         |
| `github.com/lib/pq`               | Driver PostgreSQL                            |
| `github.com/twilio/twilio-go`     | SDK Twilio para envío de SMS                 |
| `firebase.google.com/go/v4`       | SDK Firebase para notificaciones push (FCM)  |
| `github.com/google/uuid`          | Generación de identificadores únicos         |
| `github.com/nats-io/nats.go`      | Cliente NATS JetStream                       |
| `github.com/joho/godotenv`        | Carga de variables de entorno desde `.env`   |
| `github.com/caarlos0/env/v10`     | Mapeo de env vars a structs                  |
| `github.com/stretchr/testify`     | Aserciones y mocks en tests                  |
