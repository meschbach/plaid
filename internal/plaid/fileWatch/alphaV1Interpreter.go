package fileWatch

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"path/filepath"
	"time"
)

type alphaV1Interpreter struct {
	runtime *runtimeState
}

func (a *alphaV1Interpreter) Create(parent context.Context, which resources.Meta, spec AlphaV1Spec, bridge *operator.KindBridgeState) (*watch, AlphaV1Status, error) {
	ctx, span := tracing.Start(parent, "fileWatch/alphaV1.Create", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	status := AlphaV1Status{
		LastChange: nil,
	}
	if len(spec.AbsolutePath) == 0 {
		now := time.Now()
		status.LastChange = &now
		_, err := a.runtime.resources.Log(ctx, which, resources.Error, "empty absolute path")
		return nil, status, err
	}
	w := &watch{
		meta:      which,
		base:      spec.AbsolutePath,
		recursive: spec.Recursive,
		bridge:    bridge,
	}
	//w.updateLastChange()
	var err error
	if filepath.IsAbs(spec.AbsolutePath) {
		err = a.runtime.registerWatcher(ctx, w.base, w)
		if err == nil {
			status.Watching = true
		}
	} else {
		_, err = a.runtime.resources.Log(ctx, which, resources.Error, "Path %q is not absolute", spec.AbsolutePath)
	}
	return w, status, err
}

func (a *alphaV1Interpreter) Update(parent context.Context, which resources.Meta, rt *watch, s AlphaV1Spec) (AlphaV1Status, error) {
	_, span := tracing.Start(parent, "fileWatch/alphaV1Interpreter.Update", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	if rt.recursive == s.Recursive && rt.base == s.AbsolutePath {
		return AlphaV1Status{
			Watching:   rt.watching,
			LastChange: rt.lastUpdated,
		}, nil
	}

	return AlphaV1Status{}, errors.New("todo -- change chase")
}

func (a *alphaV1Interpreter) fileChanged(parent context.Context, which resources.Meta, state *runtimeState, event changeEvent) error {
	remoteContext := trace.ContextWithRemoteSpanContext(parent, event.tracing)
	ctx, span := tracing.Start(parent,
		"fileWatch/alphaV1Interpreter.fileChanged",
		trace.WithAttributes(attribute.Stringer("which", which), attribute.String("file.name", event.path)),
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithLinks(trace.LinkFromContext(remoteContext)))
	defer span.End()

	var status AlphaV1Status
	exists, err := state.resources.GetStatus(ctx, which, &status)
	if err != nil || !exists {
		return err
	}

	now := time.Now()
	status.LastChange = &now
	_, err = state.resources.UpdateStatus(ctx, which, status)
	return err
}
