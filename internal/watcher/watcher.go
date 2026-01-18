package watcher

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	configPath string
	reloadChan chan struct{}
	watcher    *fsnotify.Watcher
}

func NewWatcher(configPath string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	err = fsWatcher.Add(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to watch config file: %w", err)
	}

	// dir := filepath.Dir(configPath)
	// err = fsWatcher.Add(dir)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to watch config directory: %w", err)
	// }

	return &Watcher{
		configPath: configPath,
		reloadChan: make(chan struct{}),
		watcher:    fsWatcher,
	}, nil
}

func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if filepath.Base(event.Name) == filepath.Base(w.configPath) {
						fmt.Println()
						slog.Info("config file changed, triggering reload", "file", event.Name)

						select {
						case w.reloadChan <- struct{}{}:
						default:
						}
					}
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				slog.Error("file watcher error: ", "error", err)
			}
		}
	}()
}

func (w *Watcher) ReloadChan() <-chan struct{} {
	return w.reloadChan
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}
