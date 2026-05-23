# PRD — Sistema de Feature Flags na Azion

> **Status:** Rascunho em construção durante brainstorming.
> **Última revisão:** 2026-05-23 v0.7 — auth removida; client_id pré-existente como identificador de conta.
> **Data:** 2026-05-22
> **Autor:** Isaque Bock

---

## 1. Visão geral

Sistema de feature flags com camada de leitura na edge (Azion Edge Function) e gestão centralizada (Go API + Redis), com interface administrativa via fork do Azion Console Kit. Cada conta é identificada por um `client_id` pré-existente — sem cadastro, login ou sessões no sistema. Projeto de **estudo pessoal** — foco em aprender a stack de edge da Azion na prática.

## 2. Objetivos

- Aprender a desenvolver e deployar **Azion Edge Functions** ponta a ponta.
- Construir uma **Go API** com multi-tenancy simples por `client_id` sobre Redis.
- Aprender a **forkar e customizar o Azion Console Kit** adicionando gestão de flags.
- Explorar **edge caching** com chave de cache por `client_id`.
- Resultado tangível: criar flags pela UI informando seu `client_id`, e ter um app consumindo só as flags daquela conta via `/v1/snapshot`.

## 3. Não-objetivos (escopo cortado intencionalmente)

- ❌ Multivariante (apenas boolean).
- ❌ Rollout percentual.
- ❌ Targeting por atributos de usuário final.
- ❌ Múltiplos ambientes (dev/staging/prod).
- ❌ Teams / membros / convites.
- ❌ SDK cliente. Apps consomem via HTTP direto.
- ❌ Audit log estruturado / histórico de mudanças.
- ❌ Edge SQL ou Azion KV Store. Storage exclusivamente em **Redis**.
- ❌ Autenticação / signup / login / sessões — `client_id` já existe externamente.

## 4. Arquitetura

### 4.1. Visão de componentes

```
  Apps clientes do client_id=X    Apps clientes do client_id=Y
       │ HTTPS + X-Client-Id: X        │ HTTPS + X-Client-Id: Y
       ▼                               ▼
  ┌───────────────────────────────────────────────┐
  │  Edge Function (Azion) — TypeScript           │
  │  GET /v1/snapshot                             │──────────────┐
  │  (identifica conta via X-Client-Id)           │              │
  └───────────────────────────────────────────────┘              │ HTTP
                                                                  ▼
  ┌─────────────────────────┐              ┌───────────────────────────────┐
  │  Console Kit forkado    │  Vue/Vite    │          Go API               │
  │  (UI Admin na Azion)    │─────────────>│  CRUD flags por client_id     │
  │  Informar client_id     │  client_id   │  /v1/flags + /internal/*      │
  └─────────────────────────┘              └────────────────┬──────────────┘
                                                             │ TCP
                                                             ▼
                                                        ┌──────────┐
                                                        │  Redis   │
                                                        └──────────┘
```

**Responsabilidades:**

| Componente       | Responsabilidade                                                               |
|------------------|--------------------------------------------------------------------------------|
| Edge Function    | Resolve `client_id` → snapshot da conta; cache por `client_id`                |
| Go API           | CRUD de flags escopado por `client_id`; único acesso ao Redis                  |
| Redis            | Persiste `snapshot:<client_id>` por conta                                     |
| Console Kit fork | UI para gerenciar flags de um `client_id` informado pelo usuário               |

### 4.2. Por que essa separação

- **Go API → Redis via TCP:** Go tem clientes Redis maduros com suporte a transações; sem limitação do runtime edge.
- **Edge Function → Go API via HTTP:** a Function adiciona edge caching e distribuição geográfica.
- **UI → Go API direto:** gestão é rara; chamada direta é mais simples sem rota pela edge.

## 5. Estratégia de snapshot

### 5.1. Motivação

Cliente baixa um **snapshot** de todas as flags da sua conta de uma vez, cacheia localmente, e avalia sem rede até o cache expirar. Sem polling por flag individual.

```
App (client_id=X)     Edge Function          Go API              Redis
  │                       │                    │                   │
  │─ GET /v1/snapshot ───>│                    │                   │
  │  X-Client-Id: X       │ [cache hit         │                   │
  │                       │  pra X?]           │                   │
  │                       │─ GET /internal/snapshot ─────────────> │
  │                       │  X-Client-Id: X    │── GET snapshot:X ─>│
  │                       │                    │<─ JSON ────────────│
  │                       │<─── JSON ──────────│                   │
  │<─ {flags: {...}} ─────│                    │                   │
  │                       │                    │                   │
  │ [Cache-Control expira │                    │                   │
  │  → refaz o ciclo]     │                    │                   │
```

### 5.2. Modelo de dados no Redis

| Chave                  | Tipo   | Conteúdo                                       | Propósito                          |
|------------------------|--------|------------------------------------------------|------------------------------------|
| `snapshot:<client_id>` | String | JSON serializado: schema, flags, timestamp     | Flags da conta; criado sob demanda |

Não há tabela de contas — o `client_id` é a identidade, gerenciado externamente. O snapshot é criado automaticamente na primeira escrita de flag pra aquele `client_id`.

### 5.3. Schema de `snapshot:<client_id>`

```json
{
  "schema_version": 1,
  "client_id": "client_01HXYZ",
  "generated_at": "2026-05-23T12:00:00.000Z",
  "flags": {
    "new-checkout": {
      "enabled": true,
      "description": "Ativa novo fluxo de checkout",
      "created_at": "2026-05-23T11:00:00.000Z",
      "updated_at": "2026-05-23T11:30:00.000Z"
    }
  }
}
```

### 5.4. Write path (atômico via WATCH/MULTI/EXEC)

1. `WATCH snapshot:<client_id>`
2. `GET snapshot:<client_id>` → deserializa (`null` → snapshot vazio com `client_id` e `flags: {}`).
3. Aplica mudança em memória + atualiza `generated_at`.
4. `MULTI` + `SET snapshot:<client_id> <json>` + `EXEC`.
5. Se `EXEC` retornar `nil` (conflito), refaz do passo 1. Limite de 3 retries → `503`.

## 6. Edge Function — `GET /v1/snapshot`

### 6.1. Rota

| Método | Path           | Auth                     | Resposta                             |
|--------|----------------|--------------------------|--------------------------------------|
| GET    | `/v1/snapshot` | `X-Client-Id: <id>` (obrigatório) | `{flags: {key: bool}, generated_at}` |

### 6.2. Fluxo interno

1. Lê `X-Client-Id` do header. Ausente → `400`.
2. Faz `GET <GO_API_URL>/internal/snapshot` com `X-Internal-Token` + `X-Client-Id` repassado.
3. Go API retorna o JSON completo da conta (ou `404` se nenhuma flag existe ainda).
4. Function transforma: extrai `{key: enabled}` + preserva `generated_at`.
5. Retorna com headers de cache.

### 6.3. Formato da resposta ao cliente

```json
{
  "flags": {
    "new-checkout": true,
    "dark-mode": false
  },
  "generated_at": "2026-05-23T12:00:00.000Z"
}
```

### 6.4. Cache-Control

```
Cache-Control: public, max-age=60
```

### 6.5. Edge cache da Function (por client_id)

Cache de 30s por edge node, com `X-Client-Id` incluído na chave de cache. Cada conta tem cache isolado. Worst-case de propagação de mudança: ~90s (30s edge + 60s cliente).

---

## 7. Go API — Gestão de flags

API REST em Go, único componente com acesso ao Redis.

### 7.1. Rotas

| Método | Path                    | Auth               | Resposta            | Notas                                          |
|--------|-------------------------|--------------------|---------------------|------------------------------------------------|
| GET    | `/healthz`              | ❌ pública          | `{ok: true}`        | Health check                                   |
| GET    | `/internal/snapshot`    | 🔒 X-Internal-Token + X-Client-Id | JSON completo | Chamado pela Edge Function |
| GET    | `/v1/flags`             | `X-Client-Id`      | `{flags: [...]}`    | Lista flags da conta                           |
| POST   | `/v1/flags`             | `X-Client-Id`      | flag criada (201)   | Body: `{key, enabled, description}`            |
| GET    | `/v1/flags/:key`        | `X-Client-Id`      | flag                | 404 se não existe na conta                     |
| PATCH  | `/v1/flags/:key`        | `X-Client-Id`      | flag atualizada     | Body: `{enabled?, description?}`               |
| DELETE | `/v1/flags/:key`        | `X-Client-Id`      | 204                 | Remove só da conta informada                   |

> `X-Client-Id` nas rotas `/v1/*` é o único scoping — quem envia o header gerencia as flags daquele `client_id`. Sem verificação de identidade adicional no MVP.

### 7.2. Autenticação

- **Rotas `/internal/*`:** `X-Internal-Token` (compartilhado entre Function e Go API via env vars).
- **Rotas `/v1/*`:** sem token de autenticação — o `client_id` no header é o scopingkey. Aceitável pra MVP de aprendizado.
- **`/healthz`:** público.

### 7.3. CORS

- **Methods:** `GET, POST, PATCH, DELETE, OPTIONS`
- **Headers permitidos:** `Content-Type`, `X-Client-Id`
- **Origins:** lista via env var `ALLOWED_ORIGINS`.

### 7.4. Validação de input

- `client_id` (header): não vazio, até 128 chars.
- `key` (flag): regex `^[a-z0-9][a-z0-9-_]{0,63}$`
- `description`: string, até 200 chars.
- `enabled`: boolean.

### 7.5. Códigos de erro

| Status | Caso                                  |
|--------|---------------------------------------|
| 400    | Header/body inválido ou ausente       |
| 404    | Flag não existe na conta              |
| 409    | Flag key duplicada na conta           |
| 500    | Erro inesperado                       |
| 503    | RMW falhou após 3 retries             |

## 8. UI Admin (fork do Console Kit)

### 8.1. Setup

1. Fork de `aziontech/azion-console-kit` no GitHub.
2. Clone local + branch `feature-flags`.
3. Seguir convenções existentes do kit (blocks / components / services).

### 8.2. Telas

- **Seleção de conta:** campo para informar o `client_id`. Persiste em `localStorage`. Botão "Gerenciar flags" carrega a lista.
- **Lista de flags:** tabela com `key`, `description`, toggle `enabled`, `updated_at`. Toggle faz PATCH inline.
- **Criar flag:** form com `key`, `description`, `enabled` inicial → POST.
- **Editar flag:** mesma form preenchida; só `description` e `enabled` editáveis.

### 8.3. Componentes

Reusar os Azion Blocks (PrimeVue + Tailwind):
- `InputText` pra campo de `client_id`.
- DataTable pra lista de flags.
- Toggle/Switch pro on/off.

### 8.4. Camada de service

Criar `FeatureFlagsService` em `src/services/` — passa `X-Client-Id` em todas as chamadas à Go API.

## 9. Decisões em aberto

- ✅ **Linguagem da Function:** TypeScript.
- ✅ **Deploy da Function:** Azion CLI (`azion deploy`).
- ✅ **Identificação do cliente:** `X-Client-Id` header (pré-existente, gerenciado externamente).
- ✅ **Auth:** nenhuma — `client_id` é o scoping, sem verificação de identidade.
- ✅ **Storage:** Redis próprio, acesso TCP direto pela Go API.
- ✅ **Onde rodar a UI:** deploy na própria Azion.
- 🟡 **Onde hospedar a Go API e o Redis:** Railway, Fly.io, VPS, ou outro?

## 10. Riscos e trade-offs assumidos

- **Sem auth nas rotas `/v1/*`:** qualquer pessoa que conheça um `client_id` e a URL da Go API pode gerenciar flags daquela conta. Aceitável pro MVP de aprendizado; em produção adicionaria verificação de identidade.
- **Defasagem do cache:** ~90s worst-case pra propagação de uma mudança (30s edge cache + 60s client cache).
- **Hard delete:** flag deletada some pros clientes; comportamento do app deve ter default seguro.
- **Sem audit log:** mudanças têm só `updated_at`.

## 11. Critérios de pronto (MVP)

- [ ] Redis rodando e Go API conectada.
- [ ] Go API respondendo `/healthz`, `/v1/flags*`, `/internal/snapshot` com scoping por `client_id`.
- [ ] Edge Function respondendo `GET /v1/snapshot` com edge cache por `client_id`.
- [ ] Console Kit forkado rodando na Azion com tela de seleção de `client_id` + gestão de flags.
- [ ] Fluxo end-to-end: informar `client_id` na UI → criar flag → app chama `/v1/snapshot` com `X-Client-Id` → recebe `{flags: {"<key>": true}}` → toggle → após TTL, snapshot reflete mudança.
- [ ] README com setup + deploy dos 4 componentes.

---

## Changelog deste documento

- **2026-05-22 v0.1** — Rascunho inicial; arquitetura KV-only definida.
- **2026-05-22 v0.2** — Estratégia de snapshot; `/evaluate/:key` removido.
- **2026-05-22 v0.3** — Storage alterado de Azion KV Store para Redis.
- **2026-05-22 v0.4** — Go API adicionada; Function reduzida a snapshot público.
- **2026-05-22 v0.5** — Auditoria: `Cache-Control`, edge cache, RMW atômico, CORS, `/healthz`, `X-Internal-Token`, `schema_version`.
- **2026-05-22 v0.6** — Multi-tenancy via contas com email/senha e API key.
- **2026-05-23 v0.7** — Auth removida; `client_id` pré-existente substitui todo o sistema de contas; modelo Redis simplificado para `snapshot:<client_id>`.
