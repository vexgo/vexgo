# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build
- Full build (frontend + binary): `make build` (runs `make frontend-build` then `go build -o go-cms ./cmd/server/`)
- Frontend only: `make frontend-build` (cd admin-frontend && pnpm install && pnpm build, then copies to internal/admin/dist/)
- Manual Go build: `go build -o go-cms ./cmd/server/`
- Output binary: `go-cms` (gitignored)

### Frontend Development
- Dev server: `cd admin-frontend && pnpm dev` (http://localhost:5173, base path `/admin/`)
- Build: `cd admin-frontend && pnpm build` (outputs to admin-frontend/dist/)
- Preview: `cd admin-frontend && pnpm preview`
- Package manager: pnpm (React 18 + Vite + React Router v6)

### Test
- All tests: `make test` or `go test ./... -v`
- Single test: `go test ./<package-path> -run <TestFunctionName> -v`
- Test file example: `cmd/server/main_test.go`

### Local Run & Verification
- Build and run in background: `make run`
- Full cycle: `make verify` (build, run, sleep 2, curl http://localhost:8080/ping, kill)

### Cleanup
- `make clean` — removes `go-cms` binary

### Environment Variables
- `CMS_DSN` — SQLite DSN (default: `./data/cms.db`)
- `CMS_JWT_SECRET` — JWT signing secret (default: `change-me-in-production`)
- `CMS_PORT` — Server port (default: `8080`)

### Default Credentials
- Username: `admin`, Password: `admin123` (created automatically on first run if no users exist)

## High-Level Architecture

go-cms is a Go-based CMS distributed as a **single self-contained binary** with no runtime dependencies beyond a SQLite database file.

### Core Stack
- Go 1.25.7+, CGO required (for `mattn/go-sqlite3`)
- HTTP: Gin framework (with quic-go indirect dep for HTTP/3)
- ORM: GORM with SQLite driver (`gorm.io/driver/sqlite`)
- Auth: JWT (`golang-jwt/jwt/v5`), 24h token expiry, bcrypt password hashing
- Frontend: React 18 + Vite + React Router v6 (admin-frontend/)
- Asset embedding: `go:embed` for default theme and admin SPA
- Filesystem watching: `fsnotify` for plugin hot-reload

### Architecture Pattern: Layered

```
HTTP Request
  → Middleware (CORS → Logger → Recovery)
    → API Handlers (internal/api/)
      → Content Services (internal/cms/) — business logic
        → Database (internal/db/ + GORM models)
```

Visitor-facing pages are server-side rendered via Go `html/template` with theme support. The admin panel is a separate SPA served at `/admin/*`.

### Key Design Decisions
- **Single binary**: All assets (themes + admin SPA) embedded via `go:embed` — no external dirs needed at runtime
- **Admin SPA embedding**: Built frontend copied from `admin-frontend/dist/` to `internal/admin/dist/` by Makefile, then embedded via `internal/admin/embed.go`
- **Theme embedding**: `internal/theme/themes/{name}/` embedded via `internal/theme/embed.go`, served at `/themes/{name}/static/`
- **Plugin system**: `.so` shared libraries loaded at runtime via `plugin` package, with `fsnotify` hot-reload
- **Soft deletes**: GORM `DeletedAt` on Post and Page models
- **SEO**: Built-in sitemap.xml, robots.txt, per-page meta/OG tags

### Package Structure

| Package | Path | Purpose |
|---------|------|---------|
| Entrypoint | `cmd/server/` | `main.go` — wiring: DB, auth, router, themes, routes |
| DB/Models | `internal/db/` | SQLite connection (`sqlite.go`), GORM models (`models.go`), AutoMigrate |
| Auth | `internal/auth/` | JWT generation/validation (`jwt.go`), Gin auth middleware (`middleware.go`) |
| API Handlers | `internal/api/` | HTTP handlers: `auth.go`, `posts.go`, `pages.go`, `media.go`, `settings.go` |
| Content Services | `internal/cms/` | Business logic: `content.go` (post/page CRUD), `media.go`, `settings.go` |
| Theme System | `internal/theme/` | Loader (`loader.go`), renderer (`renderer.go`), SEO helpers (`seo.go`), embedded themes |
| Plugin System | `internal/plugin/` | Interface (`interface.go`), `.so` loader (`loader.go`), registry (`registry.go`), fsnotify watcher (`watcher.go`) |
| Middleware | `internal/middleware/` | CORS (`cors.go`), request logger (`logger.go`), panic recovery (`recovery.go`) |
| SEO | `internal/seo/` | Sitemap XML (`sitemap.go`), robots.txt (`robots.go`) |
| Embedded Admin | `internal/admin/` | `embed.go` — `//go:embed dist` for admin SPA, served at `/admin/*` |

### Data Models (`internal/db/models.go`)

| Model | Key Fields | Notes |
|-------|------------|-------|
| `Post` | Title, Slug (unique), Content, Excerpt, Status, MetaTitle, MetaDesc, OGImage, Tags, CategoryID, AuthorID, PublishedAt | Soft delete; slug auto-generated from title |
| `Page` | Title, Slug (unique), Content, Status, MetaTitle, MetaDesc, OGImage, Template | Soft delete |
| `User` | Username (unique), Password (bcrypt hash, JSON-hidden), Email, Role (`admin`/`editor`) | |
| `Category` | Name, Slug (unique) | |
| `Media` | Filename, URL, MimeType, Size | |
| `Setting` | Key (unique), Value | Key-value site settings |
| `Plugin` | Name, File, Enabled, Config | Plugin registry |
| `Theme` | Name, Active | Theme activation tracker |

### Theme System (`internal/theme/`)

Each theme in `themes/{name}/`:
```
themes/default/    (embedded via go:embed)
├── theme.json      — metadata (name, version, author, templates list)
├── templates/      — Go html/template files (base, home, post, page, archive)
└── static/        — CSS/JS served at /themes/{name}/static/
```

- `Loader` loads themes from embedded FS (`LoadEmbeddedTheme`) or filesystem (`LoadFromFS`)
- `RenderData` struct carries `SiteName`, `MetaTitle`, `MetaDescription`, `OGImage`, `CanonicalURL`, `Content`, `ThemeName`
- Custom template functions: `json`, `default`, `MetaTags`
- Active theme set via `themeLoader.SetActive()` / `GetActive()`

Available themes: `default` (full-featured), `minimal`

### Plugin System (`internal/plugin/`)

Plugins are Go `.so` shared libraries implementing the `Plugin` interface:
```go
type Plugin interface {
    Name() string
    Init(api PluginAPI) error
    Handler() gin.HandlerFunc
    Middleare() gin.HandlerFunc  // note: typo in interface
    Cleanup() error
}
```

- `Loader` loads `.so` files from `plugins/` directory
- `Registry` manages enabled plugins and their routes
- `Watcher` uses `fsnotify` to hot-reload plugins on `.so` file changes
- Plugins register routes via `PluginAPI.RegisterRoute(method, path, handler)`
- Configurable via `Plugin` DB model (file path, enabled flag, JSON config)

### Route Map (`cmd/server/main.go`)

**Public visitor routes** (SSR via themes):
- `GET /` — home page (post list)
- `GET /posts/:slug` — single post
- `GET /pages/:slug` — single page
- `GET /archive` — post + page archive

**SEO routes**:
- `GET /sitemap.xml` — auto-generated from published posts + pages
- `GET /robots.txt` — allows all, references sitemap

**Auth**:
- `POST /api/login` — returns JWT token (24h expiry)

**Public API**:
- `GET /api/posts` — list published posts
- `GET /api/posts/:slug` — get published post
- `GET /api/pages/:slug` — get published page

**Protected API** (require `Authorization: Bearer <token>`):
- Posts: `POST /api/posts`, `PUT /api/posts/:id`, `DELETE /api/posts/:id`
- Pages: `POST /api/pages`, `PUT /api/pages/:id`, `DELETE /api/pages/:id`
- Media: `GET /api/media`, `POST /api/media/upload`, `DELETE /api/media/:id`
- Settings: `GET /api/settings`, `PUT /api/settings`

**Admin SPA**:
- `GET /admin/*` — embedded React SPA with SPA fallback to index.html

### Frontend Notes
- API client: `admin-frontend/src/api/client.js` — base URL `/api`, auto-attaches JWT, redirects on 401
- Auth storage: JWT stored in `localStorage`
- Key components: `Layout.jsx`, `Login.jsx`, `Posts.jsx`, `PostEdit.jsx` (pages/media/settings are TODO stubs)
