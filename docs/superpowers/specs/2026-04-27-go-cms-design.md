# Go CMS Design Specification

> **Status:** Approved
> **Date:** 2026-04-27
> **Author:** AI Agent (superpowers:brainstorming)

## Goal

Build a Go-based CMS that is SEO-friendly, supports themes and plugins, and allows frontend development with modern frameworks. Single-binary deployment with optional external React frontend via static files embedded in Go.

## Architecture

Hybrid architecture: Go backend serves both SSR pages (for SEO) and REST APIs (for modern frontends). The single binary embeds default theme and admin frontend via `go:embed`. External themes can be loaded from `themes/` directory at runtime. Plugins use Go's `buildmode=plugin` for `.so` hot-loading.

**Tech Stack:**
- **Backend:** Go 1.21+, Gin (HTTP framework), GORM (ORM), SQLite (database)
- **Admin Frontend:** React + Vite (embedded via go:embed)
- **External Frontend:** React static build (embedded via go:embed), no Node.js required
- **Themes:** html/template + go:embed (default), filesystem themes/ (external)
- **Plugins:** Go buildmode=plugin (.so files)
- **Auth:** JWT tokens (golang-jwt/jwt)
- **SEO:** Server-side meta/OG tag injection, sitemap.xml, robots.txt

---

## 1. System Architecture

### 1.1 Request Flow

```
Browser Request
    │
    ├─ /api/*           → Gin Router → JWT Middleware → API Handlers → JSON Response
    ├─ /admin/*         → Gin Router → JWT Middleware → Serve embedded React SPA
    ├─ /themes/*/static → Gin Router → Serve theme static files (embedded or filesystem)
    ├─ /plugins/*       → Gin Router → Serve plugin static files
    └─ /* (all others)  → Gin Router → Theme Renderer → html/template SSR → HTML Response
                                    ↓
                              SEO: inject meta/OG tags + initial data
                                    ↓
                              React hydrate (if external frontend enabled)
```

### 1.2 Directory Structure

```
go-cms/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point, wiring
├── internal/
│   ├── api/                        # REST API handlers
│   │   ├── posts.go                # POST/GET/PUT/DELETE /api/posts
│   │   ├── pages.go                # POST/GET/PUT/DELETE /api/pages
│   │   ├── media.go                # Upload/list media
│   │   ├── themes.go               # List/switch themes
│   │   ├── plugins.go              # List/enable/disable plugins
│   │   └── settings.go             # Site settings
│   ├── auth/                       # JWT authentication
│   │   ├── jwt.go                  # Generate/validate JWT
│   │   └── middleware.go           # Gin JWT middleware
│   ├── cms/                        # Core business logic
│   │   ├── content.go              # Post/Page CRUD
│   │   ├── media.go                # Media management
│   │   └── settings.go             # Site settings CRUD
│   ├── plugin/                      # Plugin system
│   │   ├── loader.go               # Plugin loading (.so)
│   │   ├── registry.go             # Plugin registry
│   │   ├── interface.go             # Plugin interface definition
│   │   └── watcher.go              # File watcher for hot-reload
│   ├── theme/                       # Theme system
│   │   ├── loader.go               # Load theme (embedded or filesystem)
│   │   ├── renderer.go             # html/template rendering + SEO injection
│   │   ├── seo.go                  # Meta/OG tag generation
│   │   └── switcher.go             # Theme switching
│   ├── seo/                         # SEO utilities
│   │   ├── sitemap.go              # sitemap.xml generation
│   │   ├── robots.go               # robots.txt generation
│   │   └── meta.go                 # Meta tag helpers
│   ├── db/                          # Database layer
│   │   ├── sqlite.go               # SQLite connection + GORM setup
│   │   └── models.go               # GORM models (Post, Page, User, etc.)
│   └── middleware/                  # HTTP middleware
│       ├── cors.go                  # CORS middleware
│       ├── logger.go                # Request logging
│       └── recovery.go             # Panic recovery
├── themes/
│   └── default/                     # Default theme (go:embed compiled)
│       ├── theme.json               # Theme metadata
│       ├── templates/               # html/template files
│       │   ├── base.html            # Base template with SEO injection point
│       │   ├── home.html            # Homepage template
│       │   ├── post.html            # Single post template
│       │   ├── page.html            # Single page template
│       │   └── archive.html        # Archive/listing template
│       └── static/                  # CSS, JS, images
│           ├── style.css
│           └── app.js               # React hydrate script
├── plugins/                         # Directory for .so plugin files
│   └── (empty, plugins added by user)
├── admin-frontend/                  # React + Vite admin SPA source
│   ├── src/
│   ├── package.json
│   └── vite.config.js              # Build to dist/ → go:embed
├── public/                          # Optional external React frontend source
│   ├── src/
│   ├── package.json
│   └── vite.config.js              # Build to dist/ → go:embed
├── go.mod
├── go.sum
└── README.md
```

---

## 2. Database Design (SQLite + GORM)

### 2.1 Models

```go
// internal/db/models.go

type Post struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Title     string    `gorm:"size:255;not null" json:"title"`
    Slug      string    `gorm:"uniqueIndex;size:255;not null" json:"slug"`
    Content   string    `gorm:"type:text" json:"content"`
    Excerpt   string    `gorm:"type:text" json:"excerpt"`
    Status    string    `gorm:"size:20;default:published" json:"status"` // draft, published
    MetaTitle  string  `gorm:"size:255" json:"meta_title"`
    MetaDesc   string  `gorm:"type:text" json:"meta_description"`
    OGImage    string  `gorm:"size:500" json:"og_image"`
    Tags      string  `gorm:"type:text" json:"tags"` // comma-separated
    CategoryID uint   `json:"category_id"`
    AuthorID   uint  `json:"author_id"`
    PublishedAt *time.Time `json:"published_at"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type Page struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    Title      string    `gorm:"size:255;not null" json:"title"`
    Slug       string    `gorm:"uniqueIndex;size:255;not null" json:"slug"`
    Content    string    `gorm:"type:text" json:"content"`
    Status     string    `gorm:"size:20;default:published" json:"status"`
    MetaTitle   string  `gorm:"size:255" json:"meta_title"`
    MetaDesc    string  `gorm:"type:text" json:"meta_description"`
    OGImage     string  `gorm:"size:500" json:"og_image"`
    Template   string    `gorm:"size:100" json:"template"` // custom template
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Username  string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
    Password  string    `gorm:"size:255;not null" json:"-"` // hashed, hidden from JSON
    Email     string    `gorm:"size:255" json:"email"`
    Role      string    `gorm:"size:20;default:editor" json:"role"` // admin, editor
    CreatedAt time.Time `json:"created_at"`
}

type Category struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Name  string `gorm:"size:100;not null" json:"name"`
    Slug  string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
}

type Media struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Filename  string    `gorm:"size:255;not null" json:"filename"`
    URL       string    `gorm:"size:500;not null" json:"url"`
    MimeType  string    `gorm:"size:100" json:"mime_type"`
    Size      int64     `json:"size"`
    CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Key   string `gorm:"uniqueIndex;size:100;not null" json:"key"`
    Value string `gorm:"type:text" json:"value"`
}

type Plugin struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name       string   `gorm:"size:100;not null" json:"name"`
    File       string   `gorm:"size:255;not null" json:"file"` // .so filename
    Enabled    bool     `gorm:"default:false" json:"enabled"`
    Config     string   `gorm:"type:text" json:"config"` // JSON config
}

type Theme struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name       string   `gorm:"size:100;not null" json:"name"`
    Active     bool     `gorm:"default:false" json:"active"`
}
```

---

## 3. Theme System

### 3.1 Design

- **Default theme** (`themes/default/`) is compiled into the binary via `//go:embed`
- **External themes** are loaded from `themes/` directory at runtime
- Theme switching via config (`setting: active_theme`) or API
- Templates use Go `html/template` with custom functions for SEO

### 3.2 Theme JSON Format

```json
// themes/default/theme.json
{
    "name": "default",
    "version": "1.0.0",
    "author": "Go CMS",
    "description": "Default Go CMS theme",
    "templates": ["base", "home", "post", "page", "archive"],
    "screenshot": "screenshot.png"
}
```

### 3.3 Theme Loader

```go
// internal/theme/loader.go (design)

// Theme represents a loaded theme
type Theme struct {
    Name       string
    ThemeJSON  ThemeJSON
    Templates  *template.Template
    StaticFS   fs.FS
    BaseDir    fs.FS
}

// Loader loads themes from embedded FS (default) or filesystem (external)
type Loader struct {
    embeddedFS fs.FS        // go:embed default theme
    themesDir  string       // filesystem path for external themes
    mu         sync.RWMutex
    active     *Theme
}

// LoadActive loads the currently active theme
func (l *Loader) LoadActive(themeName string) error

// LoadFromFS loads a theme from filesystem (external theme)
func (l *Loader) LoadFromFS(themeName string) (*Theme, error)

// LoadEmbedded loads the default embedded theme
func (l *Loader) LoadEmbedded() (*Theme, error)
```

### 3.4 SEO Injection in Templates

Base template includes injection points:
```html
<!-- themes/default/templates/base.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{.MetaTitle | default .SiteName}}</title>
    <meta name="description" content="{{.MetaDescription}}">
    {{template "og-tags" .}}
    {{template "custom-head" .}}
</head>
<body>
    {{template "content" .}}
    <script>window.__INITIAL_DATA__ = {{.InitialData | json}};</script>
    <script src="/themes/{{.ThemeName}}/static/app.js"></script>
</body>
</html>
```

---

## 4. Plugin System

### 4.1 Plugin Interface

```go
// internal/plugin/interface.go

// Plugin is the interface that all plugins must implement
type Plugin interface {
    // Name returns the plugin name
    Name() string
    // Init is called when the plugin is loaded
    Init(api PluginAPI) error
    // Handler returns an optional HTTP handler for plugin routes
    Handler() http.HandlerFunc
    // Middleware returns an optional Gin middleware
    Middleware() gin.HandlerFunc
    // Cleanup is called when the plugin is unloaded
    Cleanup() error
}

// PluginAPI provides plugins access to CMS functionality
type PluginAPI interface {
    RegisterRoute(method, path string, handler gin.HandlerFunc)
    GetSetting(key string) string
    SetSetting(key, value string) error
    DB() *gorm.DB
    Log() *log.Logger
}
```

### 4.2 Plugin Loading

```go
// internal/plugin/loader.go (design)

type Loader struct {
    plugins map[string]Plugin
    mu      sync.RWMutex
}

// LoadPlugin loads a .so file and registers the plugin
func (l *Loader) LoadPlugin(soPath string) error {
    p, err := plugin.Open(soPath)
    if err != nil { return err }
    sym, err := p.Lookup("Plugin")
    if err != nil { return err }
    pluginInstance, ok := sym.(Plugin)
    if !ok { return errors.New("not a valid Plugin") }
    return pluginInstance.Init(l.api)
}
```

### 4.3 Plugin Hot-Reload

Watch `plugins/` directory for new/changed `.so` files using `fsnotify`. When a change is detected, unload old plugin and load new version.

---

## 5. REST API Design

### 5.1 Endpoints

**Public (no auth):**
```
GET  /api/posts              # List posts (pagination, filter by tag/category)
GET  /api/posts/:slug        # Get single post
GET  /api/pages/:slug        # Get single page
GET  /api/categories         # List categories
GET  /api/sitemap.xml        # Sitemap
GET  /api/robots.txt         # Robots.txt
```

**Protected (JWT required):**
```
POST   /api/login             # Get JWT token
POST   /api/posts             # Create post
PUT    /api/posts/:id          # Update post
DELETE /api/posts/:id          # Delete post
POST   /api/pages             # Create page
PUT    /api/pages/:id          # Update page
DELETE /api/pages/:id          # Delete page
POST   /api/media/upload       # Upload media
GET    /api/media              # List media
DELETE /api/media/:id          # Delete media
GET    /api/themes             # List available themes
POST   /api/themes/switch      # Switch active theme
GET    /api/plugins            # List plugins
POST   /api/plugins/:name/enable   # Enable plugin
POST   /api/plugins/:name/disable  # Disable plugin
GET    /api/settings           # Get all settings
PUT    /api/settings           # Update settings
```

### 5.2 JWT Auth Flow

```
POST /api/login {username, password}
  → Validate against DB (bcrypt hashed password)
  → Return {token: "eyJ..."}

Subsequent requests:
  Header: Authorization: Bearer eyJ...
  → JWT middleware validates → sets c.Set("user_id", claims.UserID)
```

---

## 6. Admin Frontend (React + Vite)

### 6.1 Structure

```
admin-frontend/
├── src/
│   ├── components/            # Reusable components
│   ├── pages/                # Page components
│   │   ├── Login.jsx
│   │   ├── Dashboard.jsx
│   │   ├── Posts.jsx
│   │   ├── Pages.jsx
│   │   ├── Media.jsx
│   │   ├── Themes.jsx
│   │   └── Settings.jsx
│   ├── api/                  # API client (axios/fetch)
│   ├── hooks/                # Custom React hooks
│   ├── App.jsx
│   └── main.jsx
├── package.json
└── vite.config.js            # Build to dist/ → go:embed
```

### 6.2 Build & Embed

```js
// vite.config.js
export default {
  build: {
    outDir: '../internal/admin/dist',  // Go embeds from here
    // ...
  }
}
```

```go
// internal/admin/embed.go
//go:embed dist/*
var AdminDistFS embed.FS
```

---

## 7. External Frontend (React Static)

### 7.1 Design

- User develops React app in `public/` directory
- Build output goes to `dist/` → embedded via `go:embed` into Go binary
- Go SSR renders the initial HTML with SEO data, then React hydrates
- No Node.js required at runtime

### 7.2 Hydration Flow

```
GET /posts/hello
  → Go renders templates/base.html with:
      - <title>SEO Title</title>
      - <meta> tags injected
      - window.__INITIAL_DATA__ = {post: {...}}
      - <div id="root"><!-- server-rendered content --></div>
      - <script src="/static/app.js"></script>
  → Browser: ReactDOM.hydrateRoot(<App />, document.getElementById('root'))
  → Subsequent navigation: React Router client-side
```

---

## 8. SEO Features

### 8.1 Implemented by Go Backend

- **Meta tags**: title, description, keywords per post/page
- **Open Graph**: og:title, og:description, og:image, og:url
- **Twitter Cards**: twitter:card, twitter:title, twitter:description
- **Canonical URLs**: `<link rel="canonical" href="...">`
- **sitemap.xml**: auto-generated from published posts/pages
- **robots.txt**: configurable
- **Structured data**: JSON-LD for articles

### 8.2 Sitemap Generation

```go
// internal/seo/sitemap.go
func GenerateSitemap(posts []db.Post, pages []db.Page) (string, error) {
    // Generate XML sitemap with lastmod, changefreq, priority
}
```

---

## 9. Deployment

### 9.1 Single Binary

```bash
# Build
go build -o go-cms ./cmd/server/

# Run
./go-cms
# → SQLite DB created at ./data/cms.db (default)
# → Themes loaded from embedded default + themes/ directory
# → Admin at http://localhost:8080/admin/
# → Site at http://localhost:8080/
```

### 9.2 Configuration

Via environment variables or config file (`config.yaml`):
```yaml
server:
  port: 8080
  host: "0.0.0.0"
database:
  dsn: "./data/cms.db"
theme:
  active: "default"
  themes_dir: "./themes"
plugins:
  enabled: true
  dir: "./plugins"
jwt:
  secret: "change-me-in-production"
  expiry: "24h"
```

---

## 10. Requirements Coverage

| Requirement | Implementation | Status |
|---|---|---|
| SEO Friendly | Go SSR with meta/OG injection, sitemap.xml, robots.txt | ✅ |
| Theme Support | themes/ dir, go:embed default, hot-switch via API/config | ✅ |
| Plugin System | Go buildmode=plugin, .so hot-reload, Plugin interface | ✅ |
| Modern Frontend | React+Vite admin (embedded), React static external (embedded) | ✅ |
| Single Binary Deploy | go:embed everything, only ./go-cms needed | ✅ |
| SQLite | GORM + SQLite, single file | ✅ |
| REST API | Gin handlers, JWT auth | ✅ |
