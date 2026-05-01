package theme

import (
	"bytes"
	"fmt"
	"net/http"
)

type RenderData struct {
	SiteName        string
	MetaTitle       string
	MetaDescription string
	OGImage         string
	CanonicalURL    string
	Content         interface{}
	InitialData     interface{}
	ThemeName       string
}

func (l *Loader) Render(w http.ResponseWriter, tmplName string, data RenderData) error {
	theme := l.GetActive()
	if theme == nil {
		return fmt.Errorf("no active theme")
	}
	var buf bytes.Buffer
	if err := theme.Templates.ExecuteTemplate(&buf, tmplName+".html", data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write(buf.Bytes())
	return err
}
