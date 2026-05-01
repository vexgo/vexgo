package theme

import (
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader(nil, "./themes")
	if loader == nil {
		t.Fatal("expected loader to be non-nil")
	}
	if loader.themesDir != "./themes" {
		t.Errorf("expected themesDir ./themes, got %s", loader.themesDir)
	}
}

func TestSetAndGetActive(t *testing.T) {
	loader := NewLoader(nil, "./themes")
	theme := &Theme{Name: "test"}
	loader.SetActive(theme)
	active := loader.GetActive()
	if active == nil {
		t.Fatal("expected active theme, got nil")
	}
	if active.Name != "test" {
		t.Errorf("expected test, got %s", active.Name)
	}
}
