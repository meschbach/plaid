package filewatch

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/controllers/tooling/kit"
	"github.com/meschbach/plaid/resources"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"path/filepath"
)

type alpha1Interpreter struct {
	runtime *runtimeState
}

func (a *alpha1Interpreter) Create(parent context.Context, which resources.Meta, spec Alpha1Spec, bridge kit.Manager) (*watch, error) {
	ctx, span := tracing.Start(parent, "fileWatch/alphaV1.Create", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	w := &watch{
		meta:   which,
		base:   spec.AbsolutePath,
		bridge: bridge,
	}
	if len(spec.AbsolutePath) == 0 {
		_, err := a.runtime.resources.Log(ctx, which, resources.Error, "empty absolute Path")
		return nil, err
	}
	var err error
	if filepath.IsAbs(spec.AbsolutePath) {
		err = a.runtime.registerWatcher(ctx, w.base, w)
	} else {
		_, err = a.runtime.resources.Log(ctx, which, resources.Error, "Path %q is not absolute", spec.AbsolutePath)
	}
	return w, err
}

func (a *alpha1Interpreter) Update(parent context.Context, which resources.Meta, rt *watch, s Alpha1Spec) error {
	_, span := tracing.Start(parent, "fileWatch/alpha1Interpreter.Update", trace.WithAttributes(attribute.Stringer("which", which)))
	defer span.End()

	if rt.base == s.AbsolutePath {
		return nil
	}

	return errors.New("todo -- change chase")
}

func (a *alpha1Interpreter) Delete(ctx context.Context, which resources.Meta, rt *watch) error {
	return a.runtime.unregisterWatch(ctx, rt.base, rt)
}

func (a *alpha1Interpreter) Status(ctx context.Context, rt *watch) Alpha1Status {
	return rt.asStatus()
}
