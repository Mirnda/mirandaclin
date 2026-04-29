# mirandaclin — Backend SaaS Clínicas Odontológicas

## Visão Geral

SaaS multi-tenant para clínicas odontológicas. Cada clínica é um tenant isolado.
Funcionalidades: cadastro de usuários, agendamentos, relatórios de consultas por paciente e por dentista.

---

## Stack

| Camada       | Tecnologia                                                                 |
|--------------|----------------------------------------------------------------------------|
| Linguagem    | Go 1.26+                                                                   |
| HTTP         | `net/http` stdlib (ServeMux) — sem framework                               |
| ORM          | `gorm.io/gorm` + `gorm.io/driver/postgres`                                 |
| Cache        | Redis (`redis/go-redis/v9`) com fallback `Noop`                            |
| Auth         | JWT HS256 (`golang-jwt/jwt/v5`) — `JWT_SECRET` é chave HMAC               |
| Validação    | `go-playground/validator/v10`                                              |
| Config       | `godotenv` — carrega `.env` antes de `os.Getenv`                           |
| Log          | `go.uber.org/zap` (produção: JSON estruturado; dev: console colorido)      |
| Docs         | Swagger via `swaggo/swag` + `swaggo/http-swagger/v2` (`make swagger`)     |
| Métricas     | `prometheus/client_golang` — endpoint `/metrics`                           |
| CI/CD        | GitHub Actions                                                             |
| Container    | Docker + docker-compose                                                    |

---

## Estrutura de Pastas

```
.
├── cmd/
│   └── api/
│       ├── main.go                  # entrypoint: carrega .env, inicializa DB/cache/router
│       └── routes.go                # registerRoutes: registra todas as rotas + stack de middlewares
├── internal/
│   ├── domain/
│   │   ├── shared/
│   │   │   └── address.go           # struct Address compartilhada (embedded em Profile e Clinic)
│   │   ├── tenant/
│   │   │   └── model.go             # struct Tenant (sem handler/service — gerenciado fora da API)
│   │   ├── user/
│   │   │   ├── model.go             # struct GORM + constantes de role/scope + ScopeForRole
│   │   │   ├── repository.go        # interface Repository
│   │   │   ├── service.go           # Create, Login, GetByID; emite JWT HS256
│   │   │   ├── handler.go           # handlers HTTP + anotações swagger
│   │   │   ├── helpers.go           # funções auxiliares do domínio
│   │   │   └── user_test.go
│   │   ├── profile/
│   │   │   ├── model.go             # sem timestamps/soft-delete
│   │   │   └── repository.go
│   │   ├── clinic/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── handler.go
│   │   ├── dentist_clinic/
│   │   │   ├── model.go             # sem handler/service — usado internamente por appointment
│   │   │   └── repository.go
│   │   ├── dentist_block/
│   │   │   ├── model.go             # sem handler/service — usado internamente por appointment
│   │   │   └── repository.go
│   │   ├── appointment/
│   │   │   ├── model.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── handler.go
│   │   └── consultation/
│   │       ├── model.go
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── handler.go
│   ├── health/
│   │   └── handler.go               # Liveness (/health) e Readiness (/health/ready)
│   ├── middleware/
│   │   ├── auth.go                  # valida JWT HS256, extrai tenant_id/user_id/role/scope e injeta no ctx
│   │   ├── keys.go                  # context keys + TenantFromContext, UserIDFromContext, RoleFromContext, RequestIDFromContext
│   │   ├── scope.go                 # RequireScope — admin:* concede acesso a qualquer scope
│   │   ├── cors.go                  # CORS por lista de origens (CORS_ALLOWED_ORIGINS)
│   │   ├── security_headers.go      # headers de segurança; CSP relaxado em /swagger/; HSTS fora de dev
│   │   ├── rate_limit.go            # sliding window por IP via Redis (ou Noop)
│   │   ├── request_id.go            # gera UUID v4, injeta no ctx, retorna em X-Request-ID
│   │   └── metrics.go               # middleware Prometheus por rota
│   └── infra/
│       ├── db/
│       │   ├── gorm.go              # abre conexão GORM + pool (25 open, 5 idle, 5min lifetime); executa Migrate
│       │   └── migrations/          # arquivos SQL embutidos via //go:embed, executados em ordem lexicográfica
│       │       ├── 000_tenants.sql
│       │       ├── 001_users.sql
│       │       ├── 002_profiles.sql
│       │       ├── 003_clinics.sql
│       │       ├── 004_dentist_clinics.sql
│       │       ├── 005_dentist_blocks.sql
│       │       ├── 006_appointments.sql
│       │       └── 007_consultations.sql
│       ├── cache/
│       │   ├── cache.go             # interface Cache: Get, Set, Del, Incr, Expire
│       │   ├── redis.go
│       │   └── noop.go
│       └── repository/              # implementações concretas das interfaces de domínio
│           ├── user_repository.go
│           ├── profile_repository.go
│           ├── clinic_repository.go
│           ├── dentist_clinic_repository.go
│           ├── dentist_block_repository.go
│           ├── appointment_repository.go
│           └── consultation_repository.go
├── pkg/
│   ├── config/
│   │   └── config.go                # lê variáveis de ambiente; valida JWT_SECRET obrigatório e DB_SSLMODE em prod
│   ├── logger/
│   │   ├── logger.go                # interface Logger + alias Field
│   │   └── zap.go                   # implementação zap + FromContext (retorna Noop se não injetado)
│   ├── response/
│   │   └── response.go              # helpers JSON: OK, Created, Error
│   └── validator/
│       └── validator.go
├── docs/                            # gerado por swag init (não editar manualmente)
├── .env                             # não versionar — adicionar ao .gitignore
├── .env.example                     # versionar — valores de exemplo sem segredos
├── Dockerfile                       # multi-stage: builder (golang:1.26-alpine) + runtime (alpine:3.21)
├── docker-compose.yml
├── Makefile
└── go.mod
```

---

## Padrão dos Domínios

Cada domínio segue a mesma estrutura em camadas:

```
model.go        → struct GORM, constantes, tipos; BeforeCreate gera UUID se nil
repository.go   → interface com métodos que aceitam *gorm.DB (suporte a tx)
service.go      → orquestra repositórios, aplica regras de negócio
handler.go      → decodifica request, chama service, retorna response
```

Domínios `dentist_clinic`, `dentist_block` e `profile` possuem apenas `model.go` + `repository.go` — são usados internamente por outros serviços, sem handler próprio.

**Convenção de repositório — recebe `*gorm.DB` para suportar transações:**

```go
type UserRepository interface {
    Create(ctx context.Context, db *gorm.DB, u *User) error
    FindByID(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) (*User, error)
    FindByEmail(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, email string) (*User, error)
    Update(ctx context.Context, db *gorm.DB, u *User) error
    Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}
```

**Transações (GORM) — ex.: criar usuário + perfil atomicamente:**

```go
err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := userRepo.Create(ctx, tx, user); err != nil {
        return err
    }
    p.UserID = user.ID
    return profileRepo.Create(ctx, tx, p)
})
```

**Scopo de tenant em todo acesso ao banco:**

```go
db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&users)
```

---

## Convenções de API

**Prefixo de todas as rotas de negócio:** `/v1/api/`

**Resposta de sucesso simples:**
```json
{ "success": true, "message": "operação realizada com sucesso" }
```

**Resposta com dados:**
```json
{ "success": true, "message": "ok", "data": { ... } }
```

**Resposta de erro:**
```json
{ "success": false, "message": "descrição do erro" }
```

**HTTP Status:**
- `200` — sucesso geral
- `201` — criado
- `400` — input inválido
- `401` — não autenticado
- `403` — sem permissão
- `404` — não encontrado
- `409` — conflito (ex: email duplicado)
- `500` — erro interno (nunca expor stack trace)

---

## Rotas Registradas

| Método | Rota | Middleware | Rate Limit |
|--------|------|------------|-----------|
| `GET` | `/swagger/` | — (apenas não-prod) | — |
| `GET` | `/metrics` | — | — |
| `GET` | `/health` | — | — |
| `GET` | `/health/ready` | — | — |
| `POST` | `/v1/api/auth/login` | — | 10 req/min por IP |
| `POST` | `/v1/api/users` | Auth | 120 req/min por IP |
| `GET` | `/v1/api/users/{id}` | Auth | 120 req/min por IP |
| `POST` | `/v1/api/clinics` | Auth | 120 req/min por IP |
| `GET` | `/v1/api/clinics` | Auth | 120 req/min por IP |
| `GET` | `/v1/api/clinics/{id}` | Auth | 120 req/min por IP |
| `DELETE` | `/v1/api/clinics/{id}` | Auth | 120 req/min por IP |
| `POST` | `/v1/api/appointments` | Auth | 120 req/min por IP |
| `GET` | `/v1/api/appointments/patient/{patient_id}` | Auth | 120 req/min por IP |
| `PATCH` | `/v1/api/appointments/{id}/cancel` | Auth | 120 req/min por IP |
| `POST` | `/v1/api/consultations` | Auth | 120 req/min por IP |
| `GET` | `/v1/api/consultations/patient/{patient_id}` | Auth | 30 req/min por IP |
| `GET` | `/v1/api/consultations/dentist/{dentist_id}` | Auth | 30 req/min por IP |

**Stack global de middlewares (aplicado sobre o mux inteiro):**
```
RequestID → SecurityHeaders → CORS → Metrics → rotas
```

---

## Multi-Tenant

Todas as tabelas de negócio possuem `tenant_id UUID NOT NULL`.

- O `tenant_id` é extraído dos claims do JWT após autenticação no middleware `auth.go`.
- **Todo** acesso ao banco deve filtrar por `tenant_id` — nunca consultar sem esse filtro.
- Não há middleware `tenant.go` separado — o `auth.go` já injeta `tenant_id` no contexto.

```go
tenantID := middleware.TenantFromContext(ctx) // retorna uuid.UUID
```

---

## Autenticação

### JWT Local (email/senha)
- Hash de senha com `bcrypt` (custo 12) + salt aleatório (16 bytes hex) por usuário.
- Token assinado com **HS256** usando `JWT_SECRET` como chave HMAC.
- Claims: `sub` (user_id), `tenant_id`, `role`, `scope`, `exp` (1h), `iat`.
- Auth middleware valida `SigningMethodHMAC` — rejeita qualquer outro algoritmo.

```go
// user/service.go — emissão do token
claims := jwt.MapClaims{
    "sub":       u.ID.String(),
    "tenant_id": u.TenantID.String(),
    "role":      u.Role,
    "scope":     ScopeForRole(u.Role),
    "exp":       time.Now().Add(time.Hour).Unix(),
    "iat":       time.Now().Unix(),
}
jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
```

### Scopes por role
- `dentist:read dentist:write` — dentistas
- `patient:read` — pacientes
- `admin:*` — administradores (concede acesso a qualquer scope via `RequireScope`)
- Secretary: sem scope padrão definido (retorna string vazia em `ScopeForRole`)

---

## Domínios e Modelos

### tenants

```go
type Tenant struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
    Name      string         `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### users

```go
type User struct {
    ID                    uuid.UUID      `gorm:"type:uuid;primaryKey"`
    TenantID              uuid.UUID      `gorm:"type:uuid;not null;index"`
    Email                 string         `gorm:"not null;uniqueIndex:udx_tenant_email,priority:2"`
    PasswordHash          string         `json:"-"`
    Salt                  string         `json:"-"`
    Role                  string         // admin | dentist | secretary | patient
    Phone                 string
    HasWhatsapp           bool           `gorm:"default:false"`
    EmergencyContactName  string
    EmergencyContactPhone string
    CreatedAt             time.Time
    UpdatedAt             time.Time
    DeletedAt             gorm.DeletedAt `gorm:"index"`
}
```

Índice único em banco: `(tenant_id, email)` — `udx_tenant_email`.

### profiles

Criado em transação junto com `users`. **Sem timestamps nem soft delete.**

```go
type Profile struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"`
    TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
    FullName  string
    Document  string         // CPF
    BirthDate *time.Time
    Address   shared.Address `gorm:"embedded;embeddedPrefix:address_"`
}
```

### shared.Address

Struct compartilhada em `internal/domain/shared/address.go`, embedded em `Profile` e `Clinic` com prefixo `address_`.

```go
type Address struct {
    PostalCode   string
    Street       string
    Number       string
    Complement   string
    Neighborhood string
    City         string
    State        string
    Country      string
}
```

### clinics

```go
type Clinic struct {
    ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
    TenantID      uuid.UUID      `gorm:"type:uuid;not null;index"`
    Name          string         `gorm:"not null"`
    Phone         string
    Address       shared.Address `gorm:"embedded;embeddedPrefix:address_"`
    OperatingDays pq.StringArray `gorm:"type:text[]"` // "monday"..."sunday"
    OpenTime      string         // "08:00"
    CloseTime     string         // "18:00"
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`
}
```

### dentist_clinics

```go
type DentistClinic struct {
    ID                  uuid.UUID      `gorm:"type:uuid;primaryKey"`
    TenantID            uuid.UUID      `gorm:"type:uuid;not null;index"`
    DentistID           uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:udx_dentist_clinic,priority:1"`
    ClinicID            uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:udx_dentist_clinic,priority:2"`
    WorkingDays         pq.StringArray `gorm:"type:text[]"`
    StartTime           string         // "08:00"
    EndTime             string         // "17:00"
    SlotDurationMinutes int            `gorm:"default:30"`
    Active              bool           `gorm:"default:true"`
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
// Unique constraint em banco: (dentist_id, clinic_id)
```

### dentist_blocks

```go
type DentistBlock struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
    TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
    DentistID   uuid.UUID  `gorm:"type:uuid;not null;index"`
    ClinicID    *uuid.UUID `gorm:"type:uuid"`        // nil = bloqueia em todas as clínicas
    BlockedDate time.Time  `gorm:"type:date;not null"`
    StartTime   *string                               // nil = dia inteiro
    EndTime     *string
    Reason      string
    CreatedAt   time.Time
}
```

### appointments

```go
type Appointment struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
    TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
    PatientID   uuid.UUID  `gorm:"type:uuid;not null"`
    DentistID   uuid.UUID  `gorm:"type:uuid;not null"`
    ClinicID    uuid.UUID  `gorm:"type:uuid;not null"`
    SecretaryID *uuid.UUID `gorm:"type:uuid"`
    ScheduledAt time.Time  `gorm:"not null"`
    CanceledAt  *time.Time
    Status      string     `gorm:"not null;default:scheduled"` // scheduled | completed | cancelled
    Notes       string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Constantes:** `StatusScheduled`, `StatusCompleted`, `StatusCancelled`.

**Regra de agendamento:** antes de criar um `Appointment`, verificar:
1. O par `(dentist_id, clinic_id)` existe e está ativo em `dentist_clinics`.
2. O `scheduled_at` cai em um `working_days` do dentista nessa clínica.
3. Não existe `dentist_block` cobrindo o horário solicitado.

### consultations

```go
type Consultation struct {
    ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
    TenantID      uuid.UUID `gorm:"type:uuid;not null;index"`
    AppointmentID uuid.UUID `gorm:"type:uuid;not null"`
    PatientID     uuid.UUID `gorm:"type:uuid;not null;index"`
    DentistID     uuid.UUID `gorm:"type:uuid;not null;index"`
    Diagnosis     string
    Treatment     string
    CreatedAt     time.Time
}
```

---

## Migrations

Arquivos `.sql` em `internal/infra/db/migrations/`, embutidos via `//go:embed` e executados em ordem lexicográfica (000 → 007) na inicialização.

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS
```

Cada arquivo usa `CREATE TABLE IF NOT EXISTS` — idempotente, seguro para re-executar.

---

## Cache Redis

```go
// Interface em internal/infra/cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Incr(ctx context.Context, key string) (int64, error)
    Expire(ctx context.Context, key string, ttl time.Duration) error
}

// Padrão de chave: {tenant_id}:{entidade}:{id}
key := fmt.Sprintf("%s:user:%s", tenantID, userID)

// TTL padrão:
// - sessões JWT:         1h
// - perfis:             5min
// - agendamentos do dia: 2min
// - disponibilidade:    1min
```

Nunca importar `redis` diretamente fora de `internal/infra/cache/`. Fallback `Noop` é ativado quando Redis está indisponível.

Sempre invalidar cache no `Update` e `Delete`.

---

## Testes

### Unitários
- Mockar dependências via interfaces (`Repository`, `Cache`).
- Arquivo: `{nome}_test.go` ao lado do arquivo testado.
- Cobertura mínima em services e handlers.

### Integração
- Usar PostgreSQL real — evitar divergência com comportamento de prod.
- Subir banco via `docker-compose` ou TestContainers.
- Arquivo: `{nome}_integration_test.go` com build tag `//go:build integration`.

### Executar
```bash
make test                        # unitários
go test -tags=integration ./...  # integração
```

---

## Swagger

Anotar todos os handlers antes do PR:

```go
// @Summary     Criar agendamento
// @Tags        appointments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body CreateAppointmentRequest true "Dados do agendamento"
// @Success     201 {object} Response{data=Appointment}
// @Failure     400 {object} Response
// @Router      /v1/api/appointments [post]
func (h *Handler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
```

```bash
make swagger   # swag init -g cmd/api/main.go -o docs
```

UI disponível em `/swagger/` apenas em ambientes não-produção. CSP é relaxado automaticamente para essa rota.

---

## Variáveis de Ambiente

Carregadas de `.env` via `godotenv.Load()` com fallback silencioso (`_ = godotenv.Load()`).
Versionar apenas `.env.example`; nunca versionar `.env`.

```env
APP_NAME=mirandaclin
APP_PORT=8080
APP_ENV=development          # development | stage | production

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=secret
DB_NAME=mirandaclin
DB_SSLMODE=disable           # obrigatório 'require' ou 'verify-full' em production

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=chave-hmac-secreta   # HMAC HS256 — obrigatório, falha na inicialização se vazio
JWT_ISSUER=
JWT_JWKS_URL=

CORS_ALLOWED_ORIGINS=https://app.mirandaclin.com.br,https://admin.mirandaclin.com.br
RATE_LIMIT_ENABLED=true
```

---

## Makefile

```bash
make build    # compila → bin/mirandaclin
make test     # go test -v ./...
make lint     # golangci-lint run
make swagger  # swag init -g cmd/api/main.go -o docs
make clean    # remove bin/ e docs/
```

---

## CI/CD (GitHub Actions)

Pipeline em `.github/workflows/`:
1. **lint** — `golangci-lint run`
2. **test** — `go test ./...` + integração com PostgreSQL service container
3. **build** — `docker build` + push para Amazon ECR
4. **deploy** — atualiza ECS service com a nova imagem (apenas em `main`)

---

## Deploy — AWS ECS

O serviço é implantado como container no **Amazon ECS (Elastic Container Service)**.

### Fluxo de deploy
1. GitHub Actions builda a imagem Docker e faz push para o **Amazon ECR**.
2. O step de deploy atualiza a **Task Definition** do ECS com a nova imagem.
3. O ECS Service aplica a nova Task Definition (rolling update).

### Variáveis de ambiente em produção
Em ECS, **não usar arquivo `.env`** — as variáveis são injetadas via:
- **AWS Secrets Manager** — segredos (`DB_PASS`, `JWT_SECRET`, etc.)
- **ECS Task Definition environment** — variáveis não-sensíveis (`APP_PORT`, `APP_ENV`, etc.)

### Infraestrutura esperada
| Recurso            | Descrição                                      |
|--------------------|------------------------------------------------|
| Amazon ECS         | Orquestração dos containers (Fargate ou EC2)   |
| Amazon ECR         | Registry privado das imagens Docker            |
| Amazon RDS         | PostgreSQL gerenciado                          |
| Amazon ElastiCache | Redis gerenciado                               |
| AWS Secrets Manager| Segredos injetados na Task Definition          |

### Dockerfile — multi-stage build
```
Stage 1 (builder): golang:1.26-alpine — compila binário + gera docs Swagger (apenas não-prod)
Stage 2 (runtime): alpine:3.21        — copia só o binário (~10 MB final)
```

`ARG APP_ENV` é declarado após `go mod download` para aproveitar cache do Docker.

---

## Segurança

### Rate Limiting

Middleware em `internal/middleware/rate_limit.go` — sliding window **por IP** usando Redis (ou Noop em fallback).

```go
key := fmt.Sprintf("rl:ip:%s:%s", ip, r.URL.Path)
```

**Limites por grupo:**

| Grupo                                        | Limite          | Janela |
|----------------------------------------------|-----------------|--------|
| `POST /v1/api/auth/login`                    | 10 requisições  | 1 min  |
| Rotas autenticadas (geral)                   | 120 requisições | 1 min  |
| Relatórios (`/consultations/*`)              | 30 requisições  | 1 min  |

Resposta ao exceder limite: `429 Too Many Requests` com header `Retry-After`.

---

### Security Headers

Middleware em `internal/middleware/security_headers.go` — aceita `env string` para condicionar HSTS.

```go
SecurityHeaders(cfg.AppEnv) // HSTS omitido em "development"
```

Headers aplicados em todas as rotas:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), camera=(), microphone=()`
- `Content-Security-Policy: default-src 'none'` (relaxado em `/swagger/` para inline scripts/styles)
- `Strict-Transport-Security` — apenas fora de `development`

---

### CORS

Middleware em `internal/middleware/cors.go`.

- Origens configuradas via `CORS_ALLOWED_ORIGINS` (lista separada por vírgula).
- Origens não listadas não recebem headers CORS — sem `403` explícito, browser bloqueia.
- `Access-Control-Allow-Credentials: true` apenas para origens permitidas.
- Preflight OPTIONS retorna `204 No Content`.

---

### Proteções Adicionais

- Rejeitar payloads com `Content-Type` diferente de `application/json` nos endpoints que esperam JSON.
- Limitar tamanho do body: `http.MaxBytesReader(w, r.Body, 1<<20)` (1 MB).

**Logs de segurança — registrar sempre:**
- Tentativas de login com senha inválida (sem expor motivo ao cliente).
- Tokens JWT inválidos ou expirados.
- Requisições bloqueadas por rate limit.

**Nunca registrar em log:**
- Senhas, hashes ou salts.
- Tokens JWT completos.
- Dados sensíveis de pacientes (CPF, dados clínicos).

---

## Logging

### Interface desacoplada

Nenhum pacote interno importa `zap` diretamente — todos dependem da interface `Logger` em `pkg/logger/logger.go`.

```go
type Logger interface {
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    With(fields ...Field) Logger
    Sync() error
}

type Field = zap.Field // reexporta para não vazar zap nos importadores
```

Implementação concreta em `pkg/logger/zap.go`, instanciada em `main.go`:

```go
log := logger.New(cfg.AppEnv)
defer log.Sync()
```

`logger.FromContext(ctx)` retorna o logger injetado no contexto ou um `Noop` se ausente.

### Configuração por ambiente

| `APP_ENV`     | Formato                          | Nível padrão |
|---------------|----------------------------------|--------------|
| `development` | Console colorido (human-readable) | `debug`     |
| `stage`       | JSON estruturado                 | `info`       |
| `production`  | JSON estruturado                 | `info`       |

### Campos obrigatórios por camada

```go
// Service / Repository
log.Error("falha ao criar usuário",
    zap.String("tenant_id", tenantID.String()),
    zap.Error(err),
)

// Login inválido
log.Warn("tentativa de login com senha inválida",
    zap.String("tenant_id", tenantID.String()),
)
```

`request_id` gerado pelo middleware `RequestID`, propagado via contexto e retornado em `X-Request-ID`.

---

## Observabilidade

### Métricas — `/metrics` (Prometheus)

Endpoint sem autenticação — proteger por Security Group na AWS (acesso apenas interno/VPC).

Métricas obrigatórias:

| Métrica                              | Tipo      | Labels                          |
|--------------------------------------|-----------|---------------------------------|
| `http_requests_total`                | Counter   | `method`, `path`, `status`      |
| `http_request_duration_seconds`      | Histogram | `method`, `path`                |
| `db_query_duration_seconds`          | Histogram | `operation`, `table`            |
| `cache_hits_total`                   | Counter   | `operation` (`hit`/`miss`)      |
| `rate_limit_blocked_total`           | Counter   | `route`                         |

### Health checks

```
GET /health        → liveness  (app está de pé) → 200
GET /health/ready  → readiness (DB + Redis acessíveis) → 200 ou 503
```

```json
// /health/ready — sucesso
{ "status": "ok", "checks": { "database": "ok", "cache": "ok" } }

// /health/ready — degradado
{ "status": "degraded", "checks": { "database": "unavailable", "cache": "ok" } }
```

Readiness usa timeout de 3s para ping do DB e escrita no cache.

### Request ID

`internal/middleware/request_id.go` — gera UUID v4, injeta no contexto, retorna em `X-Request-ID`.

---

## Regras Não Negociáveis

- Nunca expor stack traces ou erros internos ao cliente — logar server-side, retornar mensagem genérica.
- Todo acesso ao banco **deve** filtrar por `tenant_id`.
- Senhas e salts nunca em log, nunca em response.
- `DB_SSLMODE=disable` proibido em `production` (validado em `config.Load()`).
- `JWT_SECRET` obrigatório — `config.Load()` falha se vazio.
- Repositórios recebem `*gorm.DB` para suportar transações — nunca usar `r.db` diretamente em operações transacionais.
- Tabelas relacionais criadas em transação única (ex: `users` + `profiles`).
- Sem `panic` em handlers — usar retorno de erro + response de erro.
- Swagger atualizado a cada novo endpoint antes do PR.
- `.env` no `.gitignore`; apenas `.env.example` versionado.
- Nenhum pacote interno importa `zap` diretamente — sempre usar a interface `logger.Logger`.
- Todo log de erro deve incluir `zap.Error(err)` e o `tenant_id` quando disponível no contexto.
- Nunca logar senhas, hashes, salts, tokens JWT completos ou dados clínicos de pacientes.
- `X-Request-ID` gerado em toda requisição e propagado nos logs.
- Migrations em arquivos `.sql` separados em `internal/infra/db/migrations/` — nunca SQL inline em Go.
- Swagger via `swaggo/swag` + `swaggo/http-swagger/v2` — nunca `go-swagger` ou outra lib.
