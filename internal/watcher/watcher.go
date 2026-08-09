// Package watcher recursively watches a directory tree and invokes a
// callback after changes settle, so a burst of writes (an agent regenerating
// a whole docs folder) triggers a single re-index instead of dozens.
package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/ebenlab/byakugan/internal/scanrules"
)

// debounce is how long the tree must stay quiet before onChange fires.
const debounce = 300 * time.Millisecond

// Watcher owns the fsnotify instance and its event loop.
type Watcher struct {
	fs       *fsnotify.Watcher
	root     string
	maxDepth int
	done     chan struct{}
}

// New starts watching root and all its subdirectories, honoring the same
// skip and depth rules as the indexer (scanrules) — the depth limit also
// bounds how many kernel watches the tree consumes. onChange runs on a
// background goroutine after events settle. Newly created directories are
// added to the watch automatically.
func New(root string, onChange func()) (*Watcher, error) {
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		fs:       fs,
		root:     root,
		maxDepth: scanrules.MaxDepth(),
		done:     make(chan struct{}),
	}
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

// addTree registers watches on from and every eligible directory below it.
// Depth is always measured from the watcher's root, so directories that
// appear at runtime obey the same limit as the initial walk.
func (w *Watcher) addTree(from string) error {
	return filepath.WalkDir(from, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if path != w.root {
			rel, err := filepath.Rel(w.root, path)
			if err != nil || scanrules.SkipName(d.Name()) || scanrules.TooDeep(rel, w.maxDepth) {
				return filepath.SkipDir
			}
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
