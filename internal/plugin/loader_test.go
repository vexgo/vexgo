package plugin

import (
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("expected loader to be non-nil")
	}
}

func TestRegisterAndGetPlugin(t *testing.T) {
	loader := NewLoader()
	loader.RegisterPlugin("test-plugin", nil)
	plugin, ok := loader.GetPlugin("test-plugin")
	if !ok {
		t.Fatal("expected to find plugin")
	}
	if plugin != nil {
		t.Error("expected nil plugin value")
	}
}

func TestListPlugins(t *testing.T) {
	loader := NewLoader()
	loader.RegisterPlugin("p1", nil)
	loader.RegisterPlugin("p2", nil)
	plugins := loader.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}
}
