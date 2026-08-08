// Package watcher recursively watches a directory tree and invokes a
// callback after changes settle, so a burst of writes (an agent regenerating
// a whole docs folder) triggers a single re-index instead of dozens.
package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce is how long the tree must stay quiet before onChange fires.
const debounce = 300 * time.Millisecond

// Watcher owns the fsnotify instance and its event loop.
type Watcher struct {
	fs   *fsnotify.Watcher
	done chan struct{}
}

// New starts watching root and all its subdirectories. onChange runs on a
// background goroutine after events settle. Newly created directories are
// added to the watch automatically.
func New(root string, onChange func()) (*Watcher, error) {
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{fs: fs, done: make(chan struct{})}
	if err := w.addTree(root); err != nil {
		fs.Close()
		return nil, err
	}
	go w.loop(onChange)
	return w, nil
}

// Close stops the event loop and releases the underlying watches.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fs.Close()
}

func (w *Watcher) addTree(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != root && (strings.HasPrefix(name, ".") || name == "node_modules") {
			return filepath.SkipDir
		}
		return w.fs.Add(path)
	})
}

func (w *Watcher) loop(onChange func()) {
	var timer *time.Timer
	for {
		select {
		case <-w.done:
			return
		case ev, ok := <-w.fs.Events:
			if !ok {
				return
			}
			// Watch directories as they appear so future writes are seen.
			if ev.Op.Has(fsnotify.Create) {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					w.addTree(ev.Name)
				}
			}
			if timer == nil {
				timer = time.AfterFunc(debounce, onChange)
			} else {
				timer.Reset(debounce)
			}
		case _, ok := <-w.fs.Errors:
			if !ok {
				return
			}
		}
	}
}
