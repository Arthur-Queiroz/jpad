# jpad

Clone do Dontpad: editor de texto colaborativo minimalista, sem login. Acessar
uma URL arbitrária (`/minhanota`) salva o texto automaticamente naquele path.
Stack: Go + Chi no backend, Vue 3 no front, SQLite como armazenamento. Deploy
via Docker Compose e Cloudflare Tunnel.

## Build & Commands
Backend (de dentro de `backend/`):
- `go run ./cmd/server` — roda local
- `go test ./...` — testes
- `go build -o bin/server ./cmd/server` — build

Frontend (de dentro de `frontend/`):
- `npm run dev` — dev server
- `npm run build` — build de produção

Raiz:
- `docker compose up --build` — sobe tudo

## Project Structure
- `backend/` — API em Go (Chi + SQLite)
  - `cmd/server/` — entrypoint
  - `internal/store/` — SQLite (queries + migrations)
  - `internal/api/` — handlers Chi
- `frontend/` — Vue 3 (Vite)

## Domain Rules
- Sanitizar path: regex `^[a-zA-Z0-9_/-]+$`
- Limite de 1MB por nota
- Debounce ~500ms no front (não um PUT por tecla)
- Acessar path = upsert implícito, sem etapa de criação
- Abrir o banco pelo env `DB_PATH` com fallback local
- Erros de API retornam JSON `{ "error": "..." }` com status correto
- Migrations via `CREATE TABLE IF NOT EXISTS` no boot
- SQLite via `modernc.org/sqlite` (sem CGO); driver registra nome `"sqlite"`
- Sem ORM: usar `database/sql` puro

## API
- `GET /api/:path` → `{ content, updated_at }`
- `PUT /api/:path` → body `{ content }`, upsert
- Raiz serve o SPA; `/api` separa as rotas de dados

## Schema
`notes(path TEXT PRIMARY KEY, content TEXT NOT NULL DEFAULT '', updated_at INTEGER NOT NULL)`

## Constraints
- Sem auth no MVP
- Sem botão de salvar
- Não trocar de banco sem pedido explícito
- Não adicionar ORM
