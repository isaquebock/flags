# Go API — Arquitetura

> **Tipo:** API REST em Go.
> **Responsabilidade:** CRUD de flags escopado por `client_id`, único componente com acesso ao Redis. Atende a UI (`/v1/flags*`) e a Edge Function (`/internal/snapshot`).
> **Sem referência interna** — segue convenções idiomáticas do ecossistema Go (`cmd/` + `internal/`).

---

## 1. Stack

| Componente | Escolha | Por quê |
|------------|---------|---------|
| Linguagem | **Go 1.22+** | Sem restrições TCP (acesso direto ao Redis) |
| Router | **chi** (`github.com/go-chi/chi/v5`) | Idiomático, leve, compatível com `net/http`, middleware em árvore |
| Redis client | **go-redis/v9** (`github.com/redis/go-redis/v9`) | Cliente oficial, suporta `Watch`/`MULTI`/`EXEC` via `client.Watch()` |
| Validação | **validator/v10** + tags em structs | Padrão Go pra validação declarativa |
| Logger | **slog** (stdlib) | Sem dependência externa; JSON estruturado nativo |
| Config | **envconfig** ou stdlib `os.Getenv` | Simples, sem mágica |
| Testes | stdlib `testing` + **miniredis** | `miniredis` permite testes sem Redis real |
| Container | **Dockerfile multistage** | Deploy em Railway/Fly.io/VPS |

## 2. Estrutura de arquivos

```
api/
├── README.md                       # setup, dev, deploy
├── ARCHITECTURE.md                 # este documento
├── go.mod
├── go.sum
├── Dockerfile                      # multistage: builder + runtime distroless
├── .dockerignore
├── .env.example
├── .gitignore
├── Makefile                        # alvos: run, test, lint, build, docker
├── cmd/
│   └── api/
│       └── main.go                 # entry point: load config, wire dependencies, start server
└── internal/
    ├── config/
    │   └── config.go               # struct Config; LoadFromEnv()
    ├── server/
    │   ├── server.go               # NewServer, Run, graceful shutdown
    │   └── routes.go               # mounting de rotas + middleware chain
    ├── middleware/
    │   ├── request_id.go           # X-Request-Id por request
    │   ├── logger.go               # slog estruturado
    │   ├── recoverer.go            # recover de panics
    │   ├── cors.go                 # CORS com ALLOWED_ORIGINS
    │   ├── client_id.go            # extrai e valida X-Client-Id; injeta em request context
    │   └── internal_token.go       # valida X-Internal-Token nas rotas /internal/*
    ├── handlers/
    │   ├── health.go               # GET /healthz
    │   ├── flags.go                # CRUD /v1/flags*
    │   ├── flags_test.go
    │   ├── internal_snapshot.go    # GET /internal/snapshot
    │   └── errors.go               # writeError helper + códigos JSON
    ├── snapshot/                   # núcleo do domínio: operações sobre o snapshot
    │   ├── snapshot.go             # tipo Snapshot, FlagFull, métodos AddFlag/UpdateFlag/RemoveFlag
    │   ├── store.go                # interface Store + redisStore implementação
    │   ├── store_redis.go          # impl. com WATCH/MULTI/EXEC + retries
    │   └── store_redis_test.go     # testes com miniredis
    ├── redis/
    │   └── client.go               # NewClient(addr, password) + ping de health
    └── validation/
        └── flag.go                 # regex de key, limites de tamanho, etc.
```

## 3. Entry point (`cmd/api/main.go`)

Composição manual de dependências (sem DI container):

```go
func main() {
  cfg := config.LoadFromEnv()
  logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

  rdb, err := redisutil.NewClient(cfg.RedisAddr, cfg.RedisPassword)
  if err != nil { log.Fatalf(...) }

  store := snapshot.NewRedisStore(rdb, cfg.SnapshotMaxRetries)

  srv := server.New(server.Deps{
    Logger:        logger,
    Store:         store,
    Config:        cfg,
  })

  if err := srv.Run(ctx); err != nil { ... }
}
```

## 4. Middleware chain

```
requestId → logger → recoverer → cors → [por subrota]
  - /v1/*           → clientIdMiddleware
  - /internal/*     → internalTokenMiddleware + clientIdMiddleware
  - /healthz        → (nenhum middleware extra)
```

## 5. Rotas

| Método | Path                  | Auth                              | Handler                          |
|--------|-----------------------|-----------------------------------|----------------------------------|
| GET    | `/healthz`            | pública                           | `healthHandler`                  |
| GET    | `/internal/snapshot`  | `X-Internal-Token` + `X-Client-Id`| `internalSnapshotHandler`        |
| GET    | `/v1/flags`           | `X-Client-Id`                     | `listFlagsHandler`               |
| POST   | `/v1/flags`           | `X-Client-Id`                     | `createFlagHandler`              |
| GET    | `/v1/flags/{key}`     | `X-Client-Id`                     | `getFlagHandler`                 |
| PATCH  | `/v1/flags/{key}`     | `X-Client-Id`                     | `updateFlagHandler`              |
| DELETE | `/v1/flags/{key}`     | `X-Client-Id`                     | `deleteFlagHandler`              |

## 6. Núcleo de domínio (`internal/snapshot/`)

A peça mais importante. Encapsula o write path atômico do PRD §5.4.

### Tipo

```go
type Snapshot struct {
    SchemaVersion int                 `json:"schema_version"`
    ClientID      string              `json:"client_id"`
    GeneratedAt   *time.Time          `json:"generated_at"`
    Flags         map[string]FlagFull `json:"flags"`
}

type FlagFull struct {
    Enabled     bool      `json:"enabled"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Interface Store

```go
type Store interface {
    Get(ctx context.Context, clientID string) (*Snapshot, error)
    Mutate(ctx context.Context, clientID string, fn func(*Snapshot) error) (*Snapshot, error)
}
```

`Mutate` é o coração: aplica `fn` num snapshot lido e persiste atomicamente.

### Implementação Redis (resumo)

```go
func (s *redisStore) Mutate(ctx context.Context, clientID string, fn func(*Snapshot) error) (*Snapshot, error) {
    key := "snapshot:" + clientID
    var result *Snapshot

    err := s.client.Watch(ctx, func(tx *redis.Tx) error {
        raw, err := tx.Get(ctx, key).Result()
        snap := emptySnapshot(clientID)
        if err == nil { json.Unmarshal([]byte(raw), &snap) }
        if err != nil && !errors.Is(err, redis.Nil) { return err }

        if err := fn(snap); err != nil { return err }
        now := time.Now().UTC()
        snap.GeneratedAt = &now

        payload, _ := json.Marshal(snap)
        _, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
            return p.Set(ctx, key, payload, 0).Err()
        })
        result = snap
        return err
    }, key)

    if errors.Is(err, redis.TxFailedErr) { /* retry até maxRetries → 503 */ }
    return result, err
}
```

Handlers usam `store.Mutate(ctx, clientID, func(s *Snapshot) error { s.Flags[key] = ...; return nil })` — toda regra de negócio (já existe? duplicada?) acontece dentro do `fn`, mas garantida pelo `WATCH/MULTI/EXEC`.

## 7. Handlers (padrão)

Cada handler segue o template:

```go
func (h *flagsHandler) Create(w http.ResponseWriter, r *http.Request) {
    clientID := mw.ClientIDFromCtx(r.Context())

    var body createFlagBody
    if err := decodeAndValidate(r, &body); err != nil {
        writeJSONError(w, 400, "invalid_body", err.Error()); return
    }

    snap, err := h.store.Mutate(r.Context(), clientID, func(s *snapshot.Snapshot) error {
        if _, exists := s.Flags[body.Key]; exists {
            return errFlagExists
        }
        s.Flags[body.Key] = snapshot.FlagFull{...}
        return nil
    })

    if errors.Is(err, errFlagExists) { writeJSONError(w, 409, ...); return }
    if errors.Is(err, snapshot.ErrTooManyRetries) { writeJSONError(w, 503, ...); return }
    if err != nil { writeJSONError(w, 500, ...); return }

    writeJSON(w, 201, snap.Flags[body.Key])
}
```

## 8. Erros

Formato simples (não JSON:API pra evitar overhead):

```json
{
  "code": "flag_already_exists",
  "message": "flag with key 'new-checkout' already exists",
  "status": 409
}
```

Mapa de códigos:

| Status | Code                       |
|--------|----------------------------|
| 400    | `invalid_body`, `invalid_client_id`, `invalid_key` |
| 401    | `missing_internal_token`, `invalid_internal_token` |
| 404    | `flag_not_found`, `snapshot_not_found`             |
| 409    | `flag_already_exists`                              |
| 500    | `internal_error`                                   |
| 503    | `concurrent_write_retry_exhausted`                 |

## 9. Logging

`slog.JSONHandler` com chaves:

| Campo         | Origem                                    |
|---------------|-------------------------------------------|
| `time`        | automático                                |
| `level`       | `info`/`warn`/`error`                     |
| `msg`         | passado no log call                        |
| `request_id`  | middleware                                |
| `method`      | request                                   |
| `path`        | request                                   |
| `status`      | response writer                           |
| `duration_ms` | calculado no logger middleware             |
| `client_id`   | quando presente (não é segredo)            |

`X-Internal-Token` jamais é logado.

## 10. Config (`internal/config/config.go`)

```go
type Config struct {
    Port               string   // PORT (default 8080)
    RedisAddr          string   // REDIS_ADDR
    RedisPassword      string   // REDIS_PASSWORD
    InternalToken      string   // INTERNAL_TOKEN (compartilhado com a Function)
    AllowedOrigins     []string // ALLOWED_ORIGINS (csv)
    SnapshotMaxRetries int      // SNAPSHOT_MAX_RETRIES (default 3)
    LogLevel           string   // LOG_LEVEL (debug/info/warn/error)
}
```

`.env.example`:

```
PORT=8080
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
INTERNAL_TOKEN=change-me
ALLOWED_ORIGINS=https://console.example.com,http://localhost:5173
SNAPSHOT_MAX_RETRIES=3
LOG_LEVEL=info
```

## 11. Testes

- **`snapshot/store_redis_test.go`** — usa `miniredis` pra simular Redis em memória; testa:
  - Mutate cria snapshot do zero (chave inexistente)
  - Mutate detecta conflito (modificação concorrente entre `WATCH` e `EXEC`)
  - Retries respeitam `SnapshotMaxRetries`
  - Schema serialization round-trip
- **`handlers/flags_test.go`** — usa `httptest` + store mockado; cobre:
  - 201 em create válido
  - 409 em key duplicada
  - 404 em update/delete de key inexistente
  - 400 em body inválido / regex de key não bate
- **Sem testes E2E aqui** — ficam em diretório próprio depois.

## 12. Dockerfile (esboço)

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
```

## 13. Graceful shutdown

`server.Run(ctx)` escuta `SIGINT`/`SIGTERM`; ao receber, fecha conexões Redis e dá 10s pra requests em vôo terminarem antes de matar o processo.

## 14. Pendências / decisões locais

- 🟡 Usar `chi` ou só `net/http` + `http.ServeMux` (Go 1.22+ tem padrões mais ricos)? `chi` ganha por middleware composable.
- 🟡 Cliente HTTP global no `main` ou por handler? Default: 1 instância no `main`.
- 🟡 Deploy: container em Railway/Fly.io vs binário direto em VPS? Pendência refletida no PRD §9.
