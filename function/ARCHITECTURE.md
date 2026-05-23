# Edge Function — Arquitetura

> **Tipo:** Azion Edge Function (TypeScript, runtime Web standard / Service Worker API).
> **Responsabilidade:** servir `GET /v1/snapshot` publicamente, atuando como camada de cache global na frente da Go API.
> **Inspiração:** `dev/azion/js-azion-api-scaffold` — adotamos as mesmas convenções de entry point, env bridge, middleware chain, logger e formato de erro, mas sem auth e sem database.

---

## 1. Stack

| Componente | Escolha | Por quê |
|------------|---------|---------|
| Framework | **Hono** | Mesmo do scaffold; leve, edge-compatível, API web nativa |
| Validação | **Zod** | Mesmo do scaffold; valida `X-Client-Id` no header |
| Runtime | **Azion Edge Runtime** + Node em dev | Igual ao scaffold |
| Build | **esbuild** | Bundling pro edge; replica `build:azion` do scaffold |
| Deploy | **Azion CLI** (`azion deploy`) | Decisão já firmada no PRD |
| Testes | **Vitest** | Padrão do scaffold |
| Erros | `@azion/js-api-errors` (JSON:API) | Padrão Azion para respostas de erro |

> **Não usamos** `@azion/js-auth` aqui — a Function não autentica usuários; só repassa o header `X-Client-Id` recebido. Auth não-existe-no-MVP (ver PRD §7.2).

## 2. Estrutura de arquivos

```
function/
├── README.md                       # setup, dev, deploy
├── ARCHITECTURE.md                 # este documento
├── package.json
├── tsconfig.json
├── .env.example
├── .gitignore
├── azion/
│   ├── azion.json                  # config do Azion CLI (name, preset, function id)
│   └── args.json.example           # template de env vars passadas como FetchEvent.args
├── scripts/
│   └── deploy.sh                   # gera args.json + build + azion deploy
└── src/
    ├── azion.ts                    # LOCKED. Entry edge (addEventListener('fetch')).
    ├── azion.d.ts                  # LOCKED. Tipos do runtime Azion (FetchEvent.args).
    ├── env.ts                      # LOCKED. getEnv() — bridge args + process.env.
    ├── server.ts                   # Dev server Node (Hono via @hono/node-server).
    ├── index.ts                    # App Hono: middleware chain + rota /v1/snapshot.
    ├── config.ts                   # Config tipado: GO_API_URL, INTERNAL_TOKEN, CACHE_MAX_AGE.
    ├── types.ts                    # AppEnv + tipos compartilhados (sem AuthResult).
    ├── handlers/
    │   ├── snapshot.ts             # Handler de GET /v1/snapshot.
    │   ├── snapshot.test.ts        # Testa transformação + cache headers + erros.
    │   └── health.ts               # GET /healthz (eco simples; sem deep check).
    ├── middleware/
    │   ├── security.ts             # requestId / timeout / bodyLimit / secureHeaders.
    │   ├── validation.ts           # Zod validators (header / params).
    │   └── client-id.ts            # Lê e valida X-Client-Id; injeta em c.var.
    ├── clients/
    │   └── api-client.ts           # fetch contra a Go API (/internal/snapshot) com AbortSignal.timeout.
    └── utils/
        └── logger.ts               # Compliance logger (clientIp, clientPort, uri, etc.).
```

## 3. Entry point e runtime

Idêntico ao scaffold:

- `src/azion.ts` registra `addEventListener('fetch')`, chama `setAzionArgs(event.args)`, e delega para `app.fetch()`.
- `src/server.ts` roda o mesmo `app` localmente em Node via `@hono/node-server` pra dev.
- `src/env.ts` é o `getEnv()` que primeiro consulta `_azionArgs` (no edge) e cai pra `process.env` (em dev).

`src/index.ts` exporta o app Hono compartilhado pelas duas entradas.

## 4. Middleware chain (ordem mantida)

```
requestId → complianceLogger → timeout(30s) → bodyLimit(8KB)
  → secureHeaders → cors → [por rota] clientIdMiddleware → handler
```

Diferenças em relação ao scaffold:
- **Sem `azionAuthMiddleware`** — substituído pelo `clientIdMiddleware`.
- **`bodyLimit`** reduzido de 100KB pra 8KB — não recebemos body em `/v1/snapshot`.
- **CORS** habilitado pra permitir browsers consumirem o snapshot.

## 5. Rota `GET /v1/snapshot`

```typescript
app.get(
  '/v1/snapshot',
  clientIdMiddleware,     // valida X-Client-Id
  snapshotHandler
);
```

### Handler (resumo)

```typescript
export async function snapshotHandler(c: Context<AppEnv>) {
  const clientId = c.get('clientId');                  // injetado pelo middleware
  const internal = await fetchInternalSnapshot(clientId);

  if (internal.status === 404) return jsonApiError(c, 404, 'Snapshot not found');

  const transformed = {
    flags: mapEnabledOnly(internal.body.flags),        // {key: bool}
    generated_at: internal.body.generated_at,
  };

  return c.json(transformed, 200, {
    'Cache-Control': `public, max-age=${cfg.cacheMaxAge}`,
    'Content-Type': 'application/json',
  });
}
```

## 6. Client da Go API (`clients/api-client.ts`)

Segue o padrão de `clients/example-client.ts` do scaffold:
- `AbortSignal.timeout(cfg.apiTimeoutMs)` em toda chamada.
- Headers `X-Internal-Token` (segredo compartilhado) + `X-Client-Id` (repassado do request).
- Tipa o response: `InternalSnapshot { schema_version, client_id, generated_at, flags: Record<string, FlagFull> }`.

```typescript
export async function fetchInternalSnapshot(clientId: string): Promise<InternalResult> {
  const config = getApiConfig();
  const response = await fetch(`${config.baseUrl}/internal/snapshot`, {
    method: 'GET',
    headers: {
      'X-Internal-Token': config.internalToken,
      'X-Client-Id': clientId,
    },
    signal: AbortSignal.timeout(config.timeout),
  });
  // ...
}
```

## 7. Edge cache (Azion)

O cache do edge node é controlado por dois lados:

1. **Header `Cache-Control: public, max-age=60`** na resposta — informa proxies/CDNs.
2. **Azion Cache Rules** (via `azion.json` ou configuração no painel) — cacheia respostas por 30s **com `X-Client-Id` na chave de cache** (cache key vary). Isolamento por conta.

Documentar a regra de cache no `README.md` da function como passo manual pós-deploy (ou via JSON declarativo se a CLI suportar — verificar).

## 8. Variáveis de ambiente

```
# Acesso à Go API
GO_API_URL=https://flags-api.example.com
INTERNAL_TOKEN=<segredo compartilhado com a Go API>
API_TIMEOUT_MS=5000

# Cache
CACHE_MAX_AGE=60                  # segundos enviados no Cache-Control ao cliente

# Server (dev local)
PORT=3000
```

Em produção, populadas via `azion/args.json` (gerado pelo `scripts/deploy.sh`).

## 9. Logger e observabilidade

- Reusar o `complianceLoggerMiddleware` do scaffold integralmente — boa prática mesmo sem auth.
- Sanitização de headers sensíveis: `X-Internal-Token` adicionado à lista FORBIDDEN.
- Cada request loga `request.started` + `request.completed` com `statusCode` e `durationMs`.
- `X-Client-Id` é considerado identificador legítimo (não-sensível) — pode aparecer nos logs.

## 10. Erros

Mantém o padrão JSON:API do scaffold:

| Status | Caso |
|--------|------|
| 400 | `X-Client-Id` ausente ou fora do regex |
| 404 | Snapshot não existe para esse `client_id` |
| 504 | Timeout chamando a Go API |
| 502 | Go API retornou erro inesperado |
| 500 | Erro interno na Function |

## 11. Testes (Vitest)

- `snapshot.test.ts` — usa `app.request()` com headers mockados; mocka `fetchInternalSnapshot`; verifica:
  - resposta transformada correta (`{key: bool}`)
  - `Cache-Control` presente
  - `X-Client-Id` ausente → 400
  - timeout → 504
  - body 200 com `generated_at` preservado
- Sem testes E2E aqui — ficam num diretório separado quando todo o stack estiver up.

## 12. Pendências / decisões locais

- 🟡 Configurar Cache Rules via `azion.json` ou só via painel? (verificar suporte da CLI 2025/2026)
- 🟡 Logar `X-Client-Id` no campo `metadata.clientId` ou só no body do request? (padrão do scaffold é `metadata.clientId`)
- 🟡 Health check `/healthz` realmente útil aqui ou redundante com Azion's próprio? (incluído por padrão, custo zero)
