package mock

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"sync"
	"time"
)

// Proc represents a process which may be modified via tests.
type Proc struct {
	interpreter *alpha1Ops
	object      resources.Meta
	startedAt   *time.Time

	loggingConfig *logdrain.ServiceConfig
	StdOut        *streams.Buffer[logdrain.LogEntry]

	finishedAt *time.Time
	exitCode   int
}

func (m *Proc) Start(ctx context.Context) error {
	var err error
	gate := &sync.WaitGroup{}
	gate.Add(1)
	m.interpreter.engine.reactor.ScheduleFunc(ctx, func(ctx context.Context) error {
		now := time.Now()
		m.startedAt = &now
		err = m.interpreter.updateProc(ctx, m)
		gate.Done()
		return nil
	})
	gate.Wait()
	return err
}

func (m *Proc) Finish(ctx context.Context) error {
	var err error
	gate := &sync.WaitGroup{}
	gate.Add(1)
	m.interpreter.engine.reactor.ScheduleFunc(ctx, func(ctx context.Context) error {
		finished := time.Now()
		m.finishedAt = &finished
		m.exitCode = 0
		err = m.interpreter.updateProc(ctx, m)
		gate.Done()
		return nil
	})
	gate.Wait()
	return err
}

func (m *Proc) Write(ctx context.Context, entry logdrain.LogEntry) error {
	gate := &sync.WaitGroup{}
	gate.Add(1)
	var err error
	m.interpreter.engine.reactor.ScheduleStateFunc(ctx, func(ctx context.Context, state *engineState) error {
		err = m.StdOut.Write(ctx, entry)
		gate.Done()
		return nil
	})
	gate.Wait()
	return err
}
