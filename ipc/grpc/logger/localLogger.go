package logger

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"sync"
)

type localLogger struct {
	changes       sync.Mutex
	consumedNames map[string]streams.Sink[bufferedEntry]
}

func (l *localLogger) registerDrain(ctx context.Context, name string, output streams.Sink[bufferedEntry]) (bool, error) {
	l.changes.Lock()
	defer l.changes.Unlock()

	if _, has := l.consumedNames[name]; has {
		return false, nil
	}
	l.consumedNames[name] = output
	return true, nil
}

func (l *localLogger) unregisterDrain(ctx context.Context, name string) error {
	l.changes.Lock()
	defer l.changes.Unlock()

	if _, has := l.consumedNames[name]; !has {
		return errors.New("not registered")
	}

	delete(l.consumedNames, name)
	return nil
}

func (l *localLogger) findStream(ctx context.Context, name string) (streams.Sink[bufferedEntry], bool) {
	l.changes.Lock()
	defer l.changes.Unlock()

	sink, has := l.consumedNames[name]
	return sink, has
}
