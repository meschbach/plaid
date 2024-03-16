package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
	"sync/atomic"
)

type wireAdapterHandler struct {
	handler    resources.OnResourceChanged
	parentSpan trace.SpanContext
}

const (
	watcherAdapterReady = iota
	watcherAdapterClosed
)

// watcherAdapter is a client to fulfill watcher semantics.
type watcherAdapter struct {
	wire     reswire.ResourceControllerClient
	stream   reswire.ResourceController_WatcherClient
	tagsLock *sync.Mutex
	tags     map[resources.WatchToken]*wireAdapterHandler
	next     atomic.Int32
	//ackTable access is mediated through ackLock
	ackTable     map[uint64]*reswire.WatcherEventOut
	ackLock      *sync.Mutex
	ackCondition *sync.Cond
	//state
	state     int
	stateLock *sync.RWMutex
}

func (w *watcherAdapter) Serve(ctx context.Context) error {
	for {
		e, err := w.stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return suture.ErrDoNotRestart
			} else {
				if s, ok := status.FromError(err); ok {
					if s.Code() == codes.Canceled {
						if err := w.close(); err != nil {
							return errors.Join(err, suture.ErrDoNotRestart)
						}
						return suture.ErrDoNotRestart
					}
				}
				return err
			}
		}
		if err := w.dispatch(ctx, e); err != nil {
			return err
		}
	}
}

func (w *watcherAdapter) dispatch(serviceContext context.Context, e *reswire.WatcherEventOut) error {
	if e == nil {
		fmt.Printf("WARNING: nil event\n")
		return nil
	}
	if e.Op == reswire.WatcherEventOut_ChangeAck {
		return w.acknowledge(serviceContext, e)
	}

	if e.Ref == nil {
		fmt.Printf("WARNING: nil on ref for %#v\n", e)
		return nil
	}
	operation := reswire.InternalizeOperation(e.Op)
	which := reswire.InternalizeMeta(e.Ref)
	ctx, span := tracer.Start(serviceContext, "wire/ClientWatcher.dispatch["+operation.String()+" of "+which.Type.String()+"]")
	defer span.End()
	span.SetAttributes(which.AsTraceAttribute("which")...)
	span.SetAttributes(attribute.Int64("tag", int64(e.Tag)))
	op, has := func() (*wireAdapterHandler, bool) {
		w.tagsLock.Lock()
		defer w.tagsLock.Unlock()
		op, has := w.tags[resources.WatchToken(e.Tag)]
		return op, has
	}()
	if !has {
		span.AddEvent("tag missing")
		//todo: note tag went missing
		return nil
	}

	if err := op.handler(ctx, resources.ResourceChanged{
		Which:     which,
		Operation: operation,
		Tracing:   trace.LinkFromContext(ctx),
	}); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
			span.AddEvent("eof or canceled")
			return suture.ErrDoNotRestart
		} else {
			span.SetStatus(otelcodes.Error, "unexpected wire error")
			return err
		}
	}
	return nil
}

func (w *watcherAdapter) OnType(ctx context.Context, kind resources.Type, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	tag := w.next.Add(1)
	token := resources.WatchToken(tag)
	func() {
		w.tagsLock.Lock()
		defer w.tagsLock.Unlock()
		w.tags[token] = &wireAdapterHandler{handler: consume, parentSpan: trace.SpanContextFromContext(ctx)}
	}()

	wireTag := uint64(tag)
	k := reswire.ExternalizeType(kind)
	if err := w.emit(&reswire.WatcherEventIn{
		Tag:    wireTag,
		OnType: k,
	}); err != nil {
		return 0, err
	}
	return token, w.waitOnAck(ctx, wireTag)
}

func (w *watcherAdapter) OnResource(parent context.Context, ref resources.Meta, consume resources.OnResourceChanged) (resources.WatchToken, error) {
	ctx, span := tracer.Start(parent, "Watcher.OnResource")
	defer span.End()
	tag := w.next.Add(1)
	token := resources.WatchToken(tag)
	span.SetAttributes(attribute.Int64("tag", int64(tag)), attribute.Int64("token", int64(token)))
	func() {
		w.tagsLock.Lock()
		defer w.tagsLock.Unlock()
		w.tags[token] = &wireAdapterHandler{
			handler:    consume,
			parentSpan: trace.SpanContextFromContext(parent),
		}
	}()

	wireTag := uint64(tag)
	wireRef := reswire.MetaToWire(ref)
	if err := w.emit(&reswire.WatcherEventIn{
		Tag:        wireTag,
		OnResource: wireRef,
	}); err != nil {
		return 0, err
	}

	return token, w.waitOnAck(ctx, wireTag)
}

func noOpChangeListener(ctx context.Context, changed resources.ResourceChanged) error {
	return nil
}

func (w *watcherAdapter) Off(ctx context.Context, token resources.WatchToken) error {
	func() {
		w.ackLock.Lock()
		defer w.ackLock.Unlock()

		w.tags[token].handler = noOpChangeListener
	}()

	t := true
	return w.emit(&reswire.WatcherEventIn{
		Tag:    uint64(token),
		Delete: &t,
	})
}

func (w *watcherAdapter) emit(msg *reswire.WatcherEventIn) error {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	var err error
	switch w.state {
	case watcherAdapterReady:
		err = w.stream.Send(msg)
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok {
				if grpcStatus.Message() == "SendMsg called after CloseSend" {
					//just eat it
					return w.close()
				}
			}
		}
	default:
		//just eat the message
	}
	return err
}

func (w *watcherAdapter) Close(ctx context.Context) error {
	return w.close()
}

func (w *watcherAdapter) close() error {
	w.stateLock.Lock()
	defer w.stateLock.Unlock()

	switch w.state {
	case watcherAdapterClosed:
		//should this really be an error
		return nil
	default:
		w.state = watcherAdapterClosed
		return w.stream.CloseSend()
	}
}

func (w *watcherAdapter) waitOnAck(ctx context.Context, tag uint64) error {
	response := make(chan *reswire.WatcherEventOut, 1)
	go func() {
		w.ackLock.Lock()
		defer w.ackLock.Unlock()
		defer close(response)

		for {
			acknowledged, has := w.ackTable[tag]
			if has && acknowledged != nil {
				response <- acknowledged
				delete(w.ackTable, tag)
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
				w.ackCondition.Wait()
			}
		}
	}()
	select {
	case <-ctx.Done():
		go func() {
			w.ackLock.Lock()
			defer w.ackLock.Unlock()
			w.ackCondition.Broadcast()
		}()
		return ctx.Err()
	case <-response:
		return nil
	}
}

func (w *watcherAdapter) acknowledge(serviceContext context.Context, e *reswire.WatcherEventOut) error {
	w.ackLock.Lock()
	defer w.ackLock.Unlock()

	w.ackTable[e.Tag] = e
	w.ackCondition.Broadcast()
	return nil
}
