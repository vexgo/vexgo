package plugin

import (
	"errors"
	"plugin"
	"sync"
)

type Loader struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

func NewLoader() *Loader {
	return &Loader{
		plugins: make(map[string]Plugin),
	}
}

func (l *Loader) LoadPlugin(soPath string) (Plugin, error) {
	p, err := plugin.Open(soPath)
	if err != nil {
		return nil, errors.New("open plugin: " + err.Error())
	}
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, errors.New("lookup Plugin: " + err.Error())
	}
	pluginInstance, ok := sym.(Plugin)
	if !ok {
		return nil, errors.New("not a valid Plugin implementation")
	}
	return pluginInstance, nil
}

func (l *Loader) RegisterPlugin(name string, p Plugin) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.plugins[name] = p
}

func (l *Loader) GetPlugin(name string) (Plugin, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	p, ok := l.plugins[name]
	return p, ok
}

func (l *Loader) ListPlugins() map[string]Plugin {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make(map[string]Plugin)
	for k, v := range l.plugins {
		result[k] = v
	}
	return result
}
