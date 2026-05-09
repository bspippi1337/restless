package watch

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bspippi1337/restless/internal/events"
	"github.com/fsnotify/fsnotify"
)

func RunContext(ctx context.Context, path string, debounce time.Duration, handler func(events.Event)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := watcher.Add(abs); err != nil {
		return err
	}

	fmt.Println("watching", abs)

	last := map[string]time.Time{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			now := time.Now()
			if t, ok := last[ev.Name]; ok && now.Sub(t) < debounce {
				continue
			}
			last[ev.Name] = now

			e := events.New("fsnotify", "filesystem", ev.Name)
			e.Op = ev.Op.String()
			handler(e)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}
