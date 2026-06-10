# jpad

Clone do [Dontpad](http://dontpad.com): editor de texto colaborativo minimalista, sem login. Acesse uma URL arbitrária (`/minhanota`) e comece a escrever — o texto é salvo automaticamente.

## Stack

| Camada    | Tecnologia                                    |
|-----------|-----------------------------------------------|
| Backend   | Go + Chi                                      |
| Frontend  | Vue 3 + Vite                                  |
| Banco     | SQLite (via `modernc.org/sqlite`, sem CGO)    |
| Proxy     | Caddy (serve SPA + faz reverse p/ API)        |
| Deploy    | Docker Compose                                |

## Arquitetura

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

- Toda URL fora de `/api/*` serve o SPA Vue. O frontend lê `window.location.pathname` e faz fetch para `/api/<path>`.
- Não há rota de criação — acessar um path já faz upsert implícito.
- O salvamento é automático com debounce de 500ms no frontend.
- Limite de 1MB por nota.
- Paths são sanitizados com regex `^[a-zA-Z0-9_/-]+$`.

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

Acesse a URL que o Vite exibir (ex: http://localhost:5173). O Vite faz proxy das chamadas `/api/*` para o backend automaticamente? **Não neste setup.** Você precisará configurar o frontend para apontar ao backend ou rodar o Caddy manualmente. Para desenvolvimento rápido, edite o `vite.config.js` para adicionar proxy:

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

### Variáveis de ambiente

| Variável   | Default      | Descrição                  |
|------------|--------------|----------------------------|
| `DB_PATH`  | `notes.db`   | Caminho do arquivo SQLite  |
| `PORT`     | `8080`       | Porta do servidor Go       |
