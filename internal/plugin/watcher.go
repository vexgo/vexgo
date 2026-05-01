package plugin

import (
	"errors"
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchPlugins(pluginsDir string, loader *Loader) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write &&
					len(event.Name) > 3 && event.Name[len(event.Name)-3:] == ".so" {
					log.Printf("plugin changed: %s, reloading...", event.Name)
					// Reload plugin logic would go here
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()
	return watcher.Add(pluginsDir)
}

func (l *Loader) Stop() error {
	var firstErr error
	l.mu.Lock()
	defer l.mu.Unlock()
	for name, p := range l.plugins {
		if err := p.Cleanup(); err != nil && firstErr == nil {
			firstErr = errors.New("cleanup plugin " + name + ": " + err.Error())
		}
	}
	return firstErr
}
