package fileWatch

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"path/filepath"
)

// localFSNotify is an adapter to utilize the fsnotify library with local os related system calls for things like
// recursive watches
type localFSNotify struct {
	watch       []string
	changes     chan localFSNotifyOp
	adaptedFeed chan changeEvent
}

func newLocalFSWatcher() *localFSNotify {
	return &localFSNotify{
		changes:     make(chan localFSNotifyOp, 4),
		adaptedFeed: make(chan changeEvent, 16),
	}
}

func (l *localFSNotify) Serve(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for {
		select {
		case op := <-l.changes:
			if err := op.perform(ctx, l, watcher); err != nil {
				return err
			}
		case err := <-watcher.Errors:
			return err
		case e := <-watcher.Events:
			if err := l.consumeEvent(ctx, e); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (l *localFSNotify) ListDirectories(targetPath string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && targetPath != path {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

func (l *localFSNotify) consumeEvent(ctx context.Context, e fsnotify.Event) error {
	if e.Has(fsnotify.Write) {
		select {
		case l.adaptedFeed <- changeEvent{
			path: e.Name,
			kind: fsFileModified,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (l *localFSNotify) Watch(ctx context.Context, path string) error {
	return l.rpcInvoke(ctx, func(ctx context.Context, l *localFSNotify, w *fsnotify.Watcher) error {
		l.watch = append(l.watch, path)
		return w.Add(path)
	})
}

func (l *localFSNotify) ChangeFeed() <-chan changeEvent {
	return l.adaptedFeed
}

type localFSNotifyOp interface {
	perform(ctx context.Context, l *localFSNotify, watcher *fsnotify.Watcher) error
}

type RPCInvocation struct {
	parent context.Context
}

type RPCReply struct {
	problem error
}

type localFSRPCHandler func(ctx context.Context, l *localFSNotify, watcher *fsnotify.Watcher) error
type rpcLocalFSOp struct {
	doWork localFSRPCHandler
	tell   chan RPCReply
}

func (r *rpcLocalFSOp) perform(ctx context.Context, l *localFSNotify, watcher *fsnotify.Watcher) error {
	err := r.doWork(ctx, l, watcher)
	select {
	case r.tell <- RPCReply{
		problem: err,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *localFSNotify) rpcInvoke(ctx context.Context, do localFSRPCHandler) error {
	reply := make(chan RPCReply)
	defer close(reply)
	in := &rpcLocalFSOp{
		doWork: do,
		tell:   reply,
	}
	select {
	case l.changes <- in:
		select {
		case r := <-reply:
			return r.problem
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}
