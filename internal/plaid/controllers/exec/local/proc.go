package local

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/meschbach/go-junk-bucket/sub"
	exec2 "github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/logdrain"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type proc struct {
	control sync.Mutex
	cmd     string
	wd      string
	started *time.Time
	exit    *exitStatus

	which    resources.Meta
	onChange *operator.KindBridgeState
	logging  *logdrain.ServiceConfig

	startingLink    trace.SpanContext
	supervisionTree *suture.Supervisor
	serviceToken    suture.ServiceToken
}

type exitStatus struct {
	when    time.Time
	code    int
	problem error
}

func (p *proc) Serve(parent context.Context) error {
	ctx, span := resources.TracedMessageContext(parent, p.startingLink, "proc.Serve")
	defer span.End()

	stdout := make(chan string, 16)
	stderr := make(chan string, 16)
	done := make(chan error, 1)

	procStdout := streams.NewBuffer[logdrain.LogEntry](32, streams.WithBufferTracePrefix[logdrain.LogEntry]("proc.stdout"))
	procStderr := streams.NewBuffer[logdrain.LogEntry](32, streams.WithBufferTracePrefix[logdrain.LogEntry]("proc.stderr"))

	parts := strings.Split(p.cmd, " ")
	cmd := sub.NewSubcommand(parts[0], parts[1:])
	cmd.WithOption(&sub.WorkingDir{Where: p.wd})
	pg := sub.WithProcGroup()
	cmd.WithOption(pg)
	hasTerminated := false
	defer func() {
		if hasTerminated {
			return
		}
		if err := pg.Kill(); err != nil {
			span.SetStatus(codes.Error, "failed to terminate process group")
			span.RecordError(err)
		}
	}()

	if err := (func() error {
		initCtx, span := tracer.Start(ctx, "proc.Init")
		defer span.End()

		p.supervisionTree.Add(&logRelay{
			config:     p.logging,
			from:       stdout,
			logBuffer:  procStdout,
			ref:        p.which,
			streamName: "stdout",
			fromSpan:   trace.SpanContextFromContext(initCtx),
		})
		p.supervisionTree.Add(&logRelay{
			config:     p.logging,
			from:       stderr,
			logBuffer:  procStderr,
			ref:        p.which,
			streamName: "stderr",
			fromSpan:   trace.SpanContextFromContext(initCtx),
		})

		func() {
			p.control.Lock()
			defer p.control.Unlock()
			now := time.Now()
			p.started = &now
		}()
		if err := p.onChange.OnResourceChange(initCtx, p.which); err != nil {
			span.SetStatus(codes.Error, "failed to set start")
			return &labeledError{doing: "starting update", underlying: err}
		}

		go func() {
			err := cmd.Run(stdout, stderr)
			done <- err
		}()
		return nil
	})(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-done:
			return func() error {
				doneCtx, span := tracer.Start(ctx, "proc.Finish", trace.WithAttributes(p.which.AsTraceAttribute("which")...))
				defer span.End()
				hasTerminated = true

				func() {
					p.control.Lock()
					defer p.control.Unlock()

					if err == nil {
						span.SetAttributes(attribute.Bool("exit.error", false), attribute.Bool("exit.normal", true))
						p.exit = &exitStatus{when: time.Now(), code: 0, problem: err}
					} else if exit, ok := err.(*exec.ExitError); ok {
						code := exit.ExitCode()
						span.SetAttributes(attribute.Bool("exit.error", true), attribute.Int("exit.code", code), attribute.Bool("exit.normal", true))
						p.exit = &exitStatus{
							when:    time.Now(),
							code:    code,
							problem: err,
						}
						err = nil
					} else {
						span.SetAttributes(attribute.Bool("exit.error", true), attribute.Bool("exit.normal", false))
						span.SetStatus(codes.Error, "unknown error")
						span.RecordError(err)
						p.exit = &exitStatus{
							when:    time.Now(),
							code:    -1,
							problem: err,
						}
					}
				}()
				if err := p.onChange.OnResourceChange(doneCtx, p.which); err != nil {
					return &labeledError{doing: "updating completed state", underlying: err}
				}
				return errors.Join(err, suture.ErrDoNotRestart)
			}()
		}
	}
}

func (p *proc) toAlphaV1Status() exec2.InvocationAlphaV1Status {
	out := exec2.InvocationAlphaV1Status{}

	p.control.Lock()
	defer p.control.Unlock()
	if p.started == nil {
		return out
	}

	out.Healthy = true
	out.Started = p.started

	if p.exit == nil {
		return out
	}
	out.Finished = &p.exit.when
	out.ExitStatus = &p.exit.code
	if p.exit.problem != nil {
		out.ExitError = p.exit.problem.Error()
	}
	return out
}
