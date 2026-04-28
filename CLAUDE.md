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
| Auth         | JWT (`golang-jwt/jwt/v5`) + JWKS OAuth2/Google (`lestrrat-go/jwx/v2`)     |
| Validação    | `go-playground/validator/v10`                                              |
| Config       | `godotenv` — carrega `.env` antes de `os.Getenv`                           |
| Log          | `go.uber.org/zap` (produção: JSON estruturado; dev: console colorido)      |
| Docs         | Swagger via `swaggo/swag` (`make swagger`)                                 |
| CI/CD        | GitHub Actions                                                             |
| Container    | Docker + docker-compose                                                    |

---

## Estrutura de Pastas

```
.
├── cmd/
│   └── api/
│       └── main.go                  # entrypoint: carrega .env, inicializa DB/cache/router
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── model.go             # struct GORM + constantes de role/scope
│   │   │   ├── repository.go        # interface Repository
│   │   │   ├── service.go           # regras de negócio
│   │   │   └── handler.go           # handlers HTTP + anotações swagger
│   │   ├── profile/
│   │   ├── clinic/
│   │   ├── dentist_clinic/
│   │   ├── dentist_block/
│   │   ├── appointment/
│   │   └── consultation/
│   ├── middleware/
│   │   ├── auth.go                  # valida JWT, injeta claims no ctx
│   │   ├── tenant.go                # extrai tenant_id do JWT, injeta no ctx
│   │   └── scope.go                 # valida escopos por role
│   └── infra/
│       ├── db/
│       │   └── gorm.go              # abre conexão GORM, AutoMigrate ou migra via SQL
│       ├── cache/
│       │   ├── redis.go
│       │   └── noop.go
│       └── repository/              # implementações concretas das interfaces de domínio
│           ├── user_repository.go
│           ├── clinic_repository.go
│           └── ...
├── pkg/
│   ├── config/
│   │   └── config.go                # lê variáveis de ambiente, valida obrigatórias
│   ├── response/
│   │   └── response.go              # helpers JSON: OK, Created, Error
│   └── validator/
│       └── validator.go
├── docs/                            # gerado por swag init (não editar manualmente)
├── .env                             # não versionar — adicionar ao .gitignore
├── .env.example                     # versionar — valores de exemplo sem segredos
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

---

## Padrão dos Domínios

Cada domínio segue a mesma estrutura em camadas:

```
model.go        → struct GORM, constantes, tipos
repository.go   → interface com métodos que aceitam *gorm.DB (suporte a tx)
service.go      → orquestra repositórios, aplica regras de negócio
handler.go      → decodifica request, chama service, retorna response
```

**Convenção de repositório — recebe `*gorm.DB` para suportar transações:**

```go
type UserRepository interface {
    Create(ctx context.Context, db *gorm.DB, u *User) error
    FindByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*User, error)
    FindByEmail(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, email string) (*User, error)
    Update(ctx context.Context, db *gorm.DB, u *User) error
    Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}
```

**Transações (GORM) — ex.: criar usuário + perfil atomicamente:**

```go
err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := userRepo.Create(ctx, tx, user); err != nil {
        return err
    }
    return profileRepo.Create(ctx, tx, profile)
})
```

**Scopo de tenant em todo acesso ao banco:**

```go
db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&users)
```

---

## Convenções de API

**Prefixo de todas as rotas:** `/v1/api/`

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

## Multi-Tenant

Todas as tabelas de negócio possuem `tenant_id UUID NOT NULL`.

- O `tenant_id` é extraído do JWT claims após autenticação.
- **Todo** acesso ao banco deve filtrar por `tenant_id` — nunca consultar sem esse filtro.
- Middleware de tenant injeta o `tenant_id` no `context.Context`.

```go
tenantID := middleware.TenantFromContext(ctx) // retorna uuid.UUID
```

---

## Autenticação

### JWT Local (email/senha)
- Hash de senha com `bcrypt` + salt aleatório por usuário.
- Token assinado com chave privada RS256.
- Claims: `sub` (user_id), `tenant_id`, `role`, `scope`.

### OAuth2 / Google Sign-In
- Valida token Google via JWKS endpoint (`JWT_JWKS_URL`).
- Após validação, cria/busca usuário local e emite JWT próprio.

### Scopes por role
- `dentist:read dentist:write` — dentistas
- `patient:read` — pacientes
- `admin:*` — administradores da clínica

---

## Domínios e Modelos

### users

```go
type User struct {
    ID                   uuid.UUID  `gorm:"type:uuid;primaryKey"`
    TenantID             uuid.UUID  `gorm:"type:uuid;not null;index"`
    Email                string     `gorm:"not null"`
    PasswordHash         string
    Salt                 string
    Role                 string     // admin | dentist | secretary | patient
    Phone                string
    HasWhatsapp          bool       `gorm:"default:false"`
    EmergencyContactName string
    EmergencyContactPhone string
    CreatedAt            time.Time
    UpdatedAt            time.Time
    DeletedAt            gorm.DeletedAt `gorm:"index"` // soft delete
}
```

### profiles
Criado em transação junto com `users`.

```go
type Profile struct {
    ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
    UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
    TenantID     uuid.UUID `gorm:"type:uuid;not null;index"`
    FullName     string
    Document     string    // CPF
    BirthDate    *time.Time
    Address      Address
}

type Address struct {
	PostalCode   string `json:"postal_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Latitude     string `json:"latitude,omitempty"`
	Longitude    string `json:"longitude,omitempty"`
}


```

### clinics

```go
type Clinic struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
    TenantID     uuid.UUID      `gorm:"type:uuid;not null;index"`
    Name         string         `gorm:"not null"`
    Phone        string
    Address      Address
    // Dias de funcionamento: "monday","tuesday","wednesday","thursday","friday","saturday","sunday"
    OperatingDays pq.StringArray `gorm:"type:text[]"`
    OpenTime      string         // "08:00"
    CloseTime     string         // "18:00"
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`
}
```

### dentist_clinics
Vínculo entre dentista e clínica com horário de trabalho padrão.

```go
type DentistClinic struct {
    ID                  uuid.UUID      `gorm:"type:uuid;primaryKey"`
    TenantID            uuid.UUID      `gorm:"type:uuid;not null;index"`
    DentistID           uuid.UUID      `gorm:"type:uuid;not null"`
    ClinicID            uuid.UUID      `gorm:"type:uuid;not null"`
    // Dias que o dentista trabalha nessa clínica
    WorkingDays         pq.StringArray `gorm:"type:text[]"`
    StartTime           string         // "08:00"
    EndTime             string         // "17:00"
    SlotDurationMinutes int            `gorm:"default:30"`
    Active              bool           `gorm:"default:true"`
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
// Unique constraint: (tenant_id, dentist_id, clinic_id)
```

### dentist_blocks
Bloqueios pontuais da agenda do dentista (fora do padrão).

```go
type DentistBlock struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
    TenantID  uuid.UUID  `gorm:"type:uuid;not null;index"`
    DentistID uuid.UUID  `gorm:"type:uuid;not null"`
    // ClinicID nil = bloqueia em todas as clínicas do dentista
    ClinicID  *uuid.UUID `gorm:"type:uuid"`
    BlockedDate time.Time `gorm:"type:date;not null"`
    // StartTime/EndTime nil = bloqueia o dia inteiro
    StartTime *string
    EndTime   *string
    Reason    string
    CreatedAt time.Time
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
    ScheduledAt time.Time
    CanceledAt  *time.Time
    Status      string     // scheduled | completed | cancelled
    Notes       string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

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
    PatientID     uuid.UUID `gorm:"type:uuid;not null"`
    DentistID     uuid.UUID `gorm:"type:uuid;not null"`
    Diagnosis     string
    Treatment     string
    CreatedAt     time.Time
}
```

---

## Cache Redis

```go
// Padrão de chave: {tenant_id}:{entidade}:{id}
key := fmt.Sprintf("%s:user:%s", tenantID, userID)

// TTL padrão:
// - sessões JWT:         1h
// - perfis:             5min
// - agendamentos do dia: 2min
// - disponibilidade:    1min
```

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

---

## Variáveis de Ambiente

Carregadas de `.env` via `godotenv.Load()` no início de `main.go`.
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

JWT_ISSUER=https://accounts.google.com
JWT_JWKS_URL=https://www.googleapis.com/oauth2/v3/certs
JWT_SECRET=sua-chave-privada-rs256
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

O `godotenv.Load()` deve ser chamado com fallback silencioso (sem erro se `.env` não existir), pois em ECS o arquivo não estará presente:

```go
// main.go
_ = godotenv.Load() // ignora erro — em ECS as vars vêm do ambiente
```

### Infraestrutura esperada
| Recurso            | Descrição                                      |
|--------------------|------------------------------------------------|
| Amazon ECS         | Orquestração dos containers (Fargate ou EC2)   |
| Amazon ECR         | Registry privado das imagens Docker            |
| Amazon RDS         | PostgreSQL gerenciado                          |
| Amazon ElastiCache | Redis gerenciado                               |
| AWS Secrets Manager| Segredos injetados na Task Definition          |

---

## Segurança

### Rate Limiting

Implementado como middleware em `internal/middleware/rate_limit.go` usando Redis como backend de contagem.

**Estratégia:** sliding window por IP + por usuário autenticado (quando JWT válido presente).

```go
// Chave por IP (pré-autenticação)
key := fmt.Sprintf("rl:ip:%s", clientIP)

// Chave por usuário autenticado (pós-autenticação)
key := fmt.Sprintf("rl:user:%s:%s", tenantID, userID)
```

**Limites por rota:**

| Grupo                        | Limite          | Janela |
|------------------------------|-----------------|--------|
| `POST /v1/api/auth/*`        | 10 requisições  | 1 min  |
| `POST /v1/api/users`         | 5 requisições   | 1 min  |
| Rotas autenticadas (geral)   | 120 requisições | 1 min  |
| Relatórios / consultas       | 30 requisições  | 1 min  |

Quando o limite é atingido, retornar `429 Too Many Requests` com header `Retry-After`.

```go
// Resposta ao exceder limite
w.Header().Set("Retry-After", "60")
response.Error(w, http.StatusTooManyRequests, "muitas requisições, tente novamente em instantes")
```

---

### Security Headers

Middleware em `internal/middleware/security_headers.go` — aplicado globalmente antes de qualquer handler.

```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "0") // browsers modernos usam CSP
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Content-Security-Policy", "default-src 'none'")
        w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
        w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
        next.ServeHTTP(w, r)
    })
}
```

`Strict-Transport-Security` só é enviado em `APP_ENV=production` — em desenvolvimento pode causar problemas com HTTP local.

---

### CORS

Middleware em `internal/middleware/cors.go`.

- Origens permitidas configuradas via variável `CORS_ALLOWED_ORIGINS` (lista separada por vírgula).
- Em `production`, nunca usar `*` — rejeitar origens não listadas com `403`.
- Métodos permitidos: `GET, POST, PUT, PATCH, DELETE, OPTIONS`.
- Headers permitidos: `Authorization, Content-Type`.
- `Access-Control-Allow-Credentials: true` apenas se a origem for explicitamente permitida.

```env
CORS_ALLOWED_ORIGINS=https://app.mirandaclin.com.br,https://admin.mirandaclin.com.br
```

---

### Proteções Adicionais

**Sanitização de input:**
- Rejeitar payloads com `Content-Type` diferente de `application/json` nos endpoints que esperam JSON.
- Limitar tamanho do body: `http.MaxBytesReader(w, r.Body, 1<<20)` (1 MB).

**Logs de segurança — registrar sempre:**
- Tentativas de login com falha (sem expor motivo detalhado ao cliente).
- Tokens JWT inválidos ou expirados (IP + timestamp).
- Requisições bloqueadas por rate limit.
- Acessos negados por `tenant_id` divergente.

**Nunca registrar em log:**
- Senhas, hashes ou salts.
- Tokens JWT completos.
- Dados sensíveis de pacientes (CPF, dados clínicos).

**Variáveis de ambiente adicionais:**

```env
CORS_ALLOWED_ORIGINS=https://app.mirandaclin.com.br
RATE_LIMIT_ENABLED=true
```

---

## Logging

### Interface desacoplada

Nenhum pacote interno importa `zap` diretamente — todos dependem da interface `Logger` definida em `pkg/logger/logger.go`. Isso permite trocar a implementação em testes sem alterar código de produção.

```go
// pkg/logger/logger.go
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

A implementação concreta fica em `pkg/logger/zap.go` e é instanciada uma vez em `main.go`:

```go
// main.go
log := logger.New(cfg.AppEnv) // "development" → console; qualquer outro → JSON
defer log.Sync()
```

Serviços, repositórios e handlers recebem `logger.Logger` via injeção de dependência — nunca via variável global.

---

### Configuração por ambiente

| `APP_ENV`     | Formato     | Nível padrão | Saída   |
|---------------|-------------|--------------|---------|
| `development` | Console colorido (human-readable) | `debug` | stdout |
| `stage`       | JSON estruturado | `info`  | stdout |
| `production`  | JSON estruturado | `info`  | stdout |

Em produção o ECS/CloudWatch coleta o stdout como JSON — não usar arquivos de log.

```go
// pkg/logger/zap.go
func New(env string) Logger {
    if env == "development" {
        z, _ := zap.NewDevelopment()
        return &zapLogger{z}
    }
    z, _ := zap.NewProduction()
    return &zapLogger{z}
}
```

---

### Campos obrigatórios por camada

Todo log deve incluir os campos de contexto relevantes:

```go
// Middleware de request — injeta no ctx e loga entrada
log.Info("request",
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.String("request_id", requestID),
    zap.String("tenant_id", tenantID.String()),
    zap.String("ip", clientIP),
)

// Service / Repository — loga erros com contexto
log.Error("falha ao criar usuário",
    zap.String("tenant_id", tenantID.String()),
    zap.Error(err),
)
```

**`request_id`** gerado no middleware de entrada (UUID v4), propagado via `context.Context` e retornado no header `X-Request-ID`.

---

### Estrutura de pastas — Logger

```
pkg/
└── logger/
    ├── logger.go     # interface Logger + tipo Field
    └── zap.go        # implementação concreta com zap
```

---

## Observabilidade

### Métricas — `/metrics` (Prometheus)

Expor endpoint `/metrics` com métricas no formato Prometheus usando `prometheus/client_golang`.

Métricas obrigatórias:

| Métrica                              | Tipo      | Labels                          |
|--------------------------------------|-----------|---------------------------------|
| `http_requests_total`                | Counter   | `method`, `path`, `status`      |
| `http_request_duration_seconds`      | Histogram | `method`, `path`                |
| `db_query_duration_seconds`          | Histogram | `operation`, `table`            |
| `cache_hits_total`                   | Counter   | `operation` (`hit`/`miss`)      |
| `rate_limit_blocked_total`           | Counter   | `route`                         |

O endpoint `/metrics` **não** passa pelos middlewares de autenticação e rate limit, mas deve ser bloqueado por Security Group na AWS (acesso apenas interno/VPC).

```go
// main.go — rota fora do grupo autenticado
mux.Handle("/metrics", promhttp.Handler())
```

---

### Health checks

```
GET /health        → liveness  (app está de pé)
GET /health/ready  → readiness (DB + Redis acessíveis)
```

```json
// /health/ready — resposta de sucesso
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "cache": "ok"
  }
}
```

Retorna `503` se qualquer dependência estiver indisponível. Usado pelo ECS como `healthCheck` na Task Definition.

---

### Request ID

Middleware `internal/middleware/request_id.go`:
- Gera UUID v4 por requisição.
- Injeta no `context.Context`.
- Devolve no header de resposta `X-Request-ID`.
- Todos os logs da requisição carregam esse ID para rastreabilidade.

---

### Estrutura de pastas — Observabilidade

```
internal/
└── middleware/
    ├── request_id.go     # gera e propaga X-Request-ID
    └── metrics.go        # middleware Prometheus por rota
pkg/
└── logger/
    ├── logger.go
    └── zap.go
```

---

## Regras Não Negociáveis

- Nunca expor stack traces ou erros internos ao cliente — logar server-side, retornar mensagem genérica.
- Todo acesso ao banco **deve** filtrar por `tenant_id`.
- Senhas e salts nunca em log, nunca em response.
- `DB_SSLMODE=disable` proibido em `production` (validar em `LoadConfig`).
- Repositórios recebem `*gorm.DB` para suportar transações — nunca usar `r.db` diretamente em operações transacionais.
- Tabelas relacionais criadas em transação única (ex: `users` + `profiles`).
- Sem `panic` em handlers — usar retorno de erro + response de erro.
- Swagger atualizado a cada novo endpoint antes do PR.
- `.env` no `.gitignore`; apenas `.env.example` versionado.
- Nenhum pacote interno importa `zap` diretamente — sempre usar a interface `logger.Logger`.
- Todo log de erro deve incluir `zap.Error(err)` e o `tenant_id` quando disponível no contexto.
- Nunca logar senhas, hashes, salts, tokens JWT completos ou dados clínicos de pacientes.
- `X-Request-ID` gerado em toda requisição e propagado nos logs.
