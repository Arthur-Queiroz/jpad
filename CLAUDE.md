# jpad

Clone do Dontpad: editor de texto colaborativo minimalista, sem login. Usuário
acessa uma URL arbitrária (`/minhanota`) e o texto é salvo automaticamente
naquele path.

## Stack
- Backend: Go + Chi
- Frontend: Vue 3 (Composition API, Vite)
- DB: SQLite via `modernc.org/sqlite` (sem CGO)
- Deploy: Docker Compose + Cloudflare Tunnel

## Estrutura
- `backend/` — API em Go
  - `cmd/server/` — main.go, bootstrap do servidor
  - `internal/store/` — camada SQLite (queries, migrations)
  - `internal/api/` — handlers Chi
- `frontend/` — Vue 3 (Vite)

## Comandos
Backend (de dentro de `backend/`):
- Rodar: `go run ./cmd/server`
- Testar: `go test ./...`
- Build: `go build -o bin/server ./cmd/server`

Frontend (de dentro de `frontend/`):
- Dev: `npm run dev`
- Build: `npm run build`

Raiz:
- Docker: `docker compose up --build`

## Regras do domínio
- Sanitizar path com regex `^[a-zA-Z0-9_/-]+$` antes de qualquer query
- Limitar conteúdo a 1MB por nota
- Debounce de ~500ms no front; não fazer um PUT por tecla
- Path não precisa ser "criado": acessar = upsert implícito
- Abrir o banco pelo env `DB_PATH` com fallback local
- Erros de API retornam JSON `{ "error": "..." }` com status adequado
- Migrations via `CREATE TABLE IF NOT EXISTS` no boot
- Driver SQLite registra o nome `"sqlite"` (não `"sqlite3"`)

## API
- `GET /api/:path` → `{ content, updated_at }` (vazio se inexistente)
- `PUT /api/:path` → body `{ content }`, faz upsert
- Rota raiz serve o SPA; prefixo `/api` evita conflito com paths de notas

## Schema
notes(path TEXT PRIMARY KEY, content TEXT NOT NULL DEFAULT '', updated_at INTEGER NOT NULL)

## Não fazer
- Não adicionar autenticação no MVP
- Não adicionar botão de salvar (auto-save sempre)
- Não trocar SQLite por outro banco sem pedido explícito
- Não adicionar ORM; usar `database/sql` puro
