package theme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

type ThemeJSON struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Templates   []string `json:"templates"`
}

type Theme struct {
	Name       string
	ThemeJSON  ThemeJSON
	Templates  *template.Template
	StaticFS   fs.FS
	BaseDir    fs.FS
}

type Loader struct {
	embeddedFS fs.FS
	themesDir  string
	mu         sync.RWMutex
	active     *Theme
}

func NewLoader(embedded fs.FS, themesDir string) *Loader {
	return &Loader{
		embeddedFS: embedded,
		themesDir:  themesDir,
	}
}

func (l *Loader) LoadEmbedded() (*Theme, error) {
	if l.embeddedFS == nil {
		return nil, fmt.Errorf("no embedded FS")
	}
	return l.LoadEmbeddedTheme("default")
}

func (l *Loader) LoadEmbeddedTheme(themeName string) (*Theme, error) {
	if l.embeddedFS == nil {
		return nil, fmt.Errorf("no embedded FS")
	}
	subFS, err := fs.Sub(l.embeddedFS, filepath.Join("themes", themeName))
	if err != nil {
		return nil, fmt.Errorf("sub FS: %w", err)
	}
	return l.loadFromFS(subFS, ".")
}

func (l *Loader) LoadFromFS(themeName string) (*Theme, error) {
	dir := filepath.Join(l.themesDir, themeName)
	dirFS := os.DirFS(dir)
	return l.loadFromFS(dirFS, ".")
}

func (l *Loader) loadFromFS(fsys fs.FS, dir string) (*Theme, error) {
	raw, err := fs.ReadFile(fsys, filepath.Join(dir, "theme.json"))
	if err != nil {
		return nil, fmt.Errorf("read theme.json: %w", err)
	}
	var tj ThemeJSON
	if err := json.Unmarshal(raw, &tj); err != nil {
		return nil, fmt.Errorf("parse theme.json: %w", err)
	}

	tmplFS, err := fs.Sub(fsys, filepath.Join(dir, "templates"))
	if err != nil {
		return nil, fmt.Errorf("open templates: %w", err)
	}
	tmpl := template.New("").Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"default": func(value, defaultValue interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
		"MetaTags": MetaTags,
	})
	tmpl, err = tmpl.ParseFS(tmplFS, "*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	staticFS, err := fs.Sub(fsys, filepath.Join(dir, "static"))
	if err != nil {
		staticFS = nil
	}

	return &Theme{
		Name:      tj.Name,
		ThemeJSON: tj,
		Templates:  tmpl,
		StaticFS:  staticFS,
		BaseDir:   fsys,
	}, nil
}

func (l *Loader) SetActive(theme *Theme) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.active = theme
}

func (l *Loader) GetActive() *Theme {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.active
}
