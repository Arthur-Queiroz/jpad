# jpad

**Clone do [Dontpad](http://dontpad.com)**: editor de texto colaborativo minimalista, sem login. Acesse uma URL arbitrária e comece a escrever — o texto é salvo automaticamente.

🌐 **Produção:** https://jpad.devarthur.com.br/

## Features

- ✍️ **Salvamento automático** com debounce de 500ms
- 🔗 **Sem login** — a URL é seu identificador
- 🌓 **Tema claro e escuro** com persistência local
- 💾 **SQLite** como armazenamento (zero configuração)
- 🚀 **Deploy em 1 comando** com Docker Compose
- ✅ **100% testado** (store + API handlers)

## Stack

| Camada    | Tecnologia                                    |
|-----------|-----------------------------------------------|
| Backend   | Go + Chi (router)                             |
| Frontend  | Vue 3 + Vite                                  |
| Banco     | SQLite (via `modernc.org/sqlite`, sem CGO)    |
| Proxy     | Caddy (serve SPA + faz reverse p/ API)        |
| Deploy    | Docker Compose                                |

## Como funciona

```
Caddy (:80)
 ├── /api/*  → reverse_proxy backend:8080
 └── /*      → file_server (SPA Vue) com fallback p/ index.html
                |
             Backend (Go :8080)
              ├── GET  /api/:path → retorna { content, updated_at }
              └── PUT  /api/:path → recebe { content }, faz upsert
                    |
                 SQLite (notes.db)
                  └── notes(path TEXT PK, content TEXT, updated_at INT)
```

### Fluxo do usuário

1. Usuário acessa `/minhanota`
2. Caddy serve o SPA Vue (fallback para `index.html`)
3. Frontend lê `window.location.pathname` e faz `GET /api/minhanota`
4. Se a nota não existe, retorna conteúdo vazio (upsert implícito)
5. Usuário digita texto
6. Frontend faz `PUT /api/minhanota` com debounce de 500ms
7. Backend faz upsert no SQLite
8. Repete do passo 5

### Regras de negócio

- **Não há rota de criação** — acessar um path já faz upsert implícito
- **Limite de 1MB por nota** — retorna `413 Request Entity Too Large`
- **Paths sanitizados** — apenas `^[a-zA-Z0-9_/-]+$` é aceito
- **Nota inexistente** — `GET` retorna `{ content: "", updated_at: 0 }` (200, não 404)
- **Upsert atômico** — `ON CONFLICT DO UPDATE` no SQLite

## Estrutura do projeto

```
jpad/
├── backend/
│   ├── cmd/server/          # Entry point (main.go)
│   ├── internal/
│   │   ├── api/             # Handlers Chi (HTTP)
│   │   └── store/           # SQLite (queries + migrations)
│   ├── Dockerfile           # Build multi-stage (Go)
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.vue          # SPA (home + editor)
│   │   ├── composables/     # useTheme.js
│   │   └── style.css        # CSS vars (temas)
│   ├── Dockerfile           # (dentro de caddy/)
│   └── package.json
├── caddy/
│   ├── Caddyfile            # Reverse proxy + SPA fallback
│   └── Dockerfile           # Build frontend + Caddy
├── docker-compose.yml       # Orquestra backend + caddy
└── README.md
```

## Rodar localmente

### Com Docker Compose (recomendado)

```bash
docker compose up --build
```

Acesse http://localhost:8080.

### Manual (sem Docker)

**Pré-requisitos:** Go 1.26+, Node.js 22+.

```bash
# Terminal 1: backend
cd backend
go run ./cmd/server

# Terminal 2: frontend
cd frontend
npm install
npm run dev
```

O Vite roda na porta `5173` por padrão, mas **precisa de proxy** para as chamadas `/api/*` chegarem ao backend. Adicione isso ao `frontend/vite.config.js`:

```js
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

Agora acesse a URL que o Vite exibir (ex: http://localhost:5173).

### Variáveis de ambiente

| Variável   | Default      | Descrição                  |
|------------|--------------|----------------------------|
| `DB_PATH`  | `notes.db`   | Caminho do arquivo SQLite  |
| `PORT`     | `8080`       | Porta do servidor Go       |

## Testes

O backend tem testes de integração para store e API handlers, usando apenas a stdlib (`testing`).

```bash
# Rodar todos os testes
cd backend
go test ./...

# Verbose
go test -v ./...

# Teste específico
go test -run TestUpsert ./...

# Com coverage
go test -cover ./...
```

## API

### `GET /api/:path`

Retorna o conteúdo de uma nota.

**Exemplo:**
```bash
curl http://localhost:8080/api/minha-nota
```

**Resposta (nota existe):**
```json
{
  "path": "minha-nota",
  "content": "texto da nota",
  "updated_at": 1718456789
}
```

**Resposta (nota não existe):**
```json
{
  "path": "minha-nota",
  "content": "",
  "updated_at": 0
}
```

### `PUT /api/:path`

Cria ou atualiza uma nota.

**Body:**
```json
{
  "content": "novo texto"
}
```

**Exemplo:**
```bash
curl -X PUT http://localhost:8080/api/minha-nota \
  -H "Content-Type: application/json" \
  -d '{"content": "novo texto"}'
```

**Resposta:**
```json
{
  "path": "minha-nota",
  "content": "novo texto",
  "updated_at": 1718456789
}
```

## Deploy

O Docker Compose sobe dois serviços:

1. **backend** — build multi-stage do Go, expõe porta 8080 internamente
2. **caddy** — build do frontend (Node), copia dist para `/srv`, Caddy serve na porta 80 (mapeada para `127.0.0.1:8081`)

Para produção, use um proxy reverso (nginx, Traefik, Cloudflare Tunnel) na frente do Caddy para adicionar HTTPS e domínio customizado.

## Licença

MIT
