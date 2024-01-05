package fileWatch

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/junk"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/meschbach/go-junk-bucket/testing/faking"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"testing"
	"time"
)

func TestNewFileWatch(t *testing.T) {
	onDone := junk.SetupTestTracing(t)
	t.Cleanup(func() {
		onDone(context.Background())
	})

	spanCtx, span := tracing.Start(context.Background(), t.Name())
	t.Cleanup(func() {
		if t.Failed() {
			span.SetStatus(codes.Error, "test failure")
		}
		span.End()
	})
	topCtx, done := context.WithTimeout(spanCtx, 5*time.Second)
	t.Cleanup(func() {
		done()
	})

	plaid := resources.WithTestSubsystem(t, topCtx)
	mockFS := NewMockFileWatcher(plaid)
	refNames := faking.DistinctWords(5)

	t.Run("Given a nonexistent file", func(t *testing.T) {
		t.Run("When asked to watch the file", func(t *testing.T) {
			ctx, span := tracing.Start(topCtx, t.Name())
			t.Cleanup(func() {
				if t.Failed() {
					span.SetStatus(codes.Error, "test failure")
				}
				span.End()
			})

			ref := resources.RequireCreate(t, ctx, plaid.Store, AlphaV1Type, refNames[0], AlphaV1Spec{AbsolutePath: "/tmp/test"})
			t.Run("Then it does not have changes", func(t *testing.T) {
				resources.AssertStatus(t, ctx, plaid.Store, ref, func(status AlphaV1Status) bool {
					return status.LastChange == nil
				})
			})
		})
	})

	t.Run("Given an existing file", func(t *testing.T) {
		fileName := "/tmp/file-exists"
		mockFS.GivenFileExists(fileName)

		t.Run("When asked to watch the file", func(t *testing.T) {
			ctx, _ := traceSubtest(t, topCtx)

			ref := resources.RequireCreate(t, ctx, plaid.Store, AlphaV1Type, refNames[1], AlphaV1Spec{AbsolutePath: fileName})

			t.Run("Then no initial change is recorded", func(t *testing.T) {
				AssertNoLastChange(t, ctx, plaid, ref)
			})

			t.Run("And the file is changed", func(t *testing.T) {
				ctx, _ = traceSubtest(t, ctx)
				changedTime := time.Now()
				mockFS.FileChanged(t, ctx, fileName)

				t.Run("Then the file records the change time", func(t *testing.T) {
					ctx, _ = traceSubtest(t, ctx)
					resources.AssertStatus(t, ctx, plaid.Store, ref, func(status AlphaV1Status) bool {
						if status.LastChange == nil {
							return false
						}
						when := *status.LastChange
						return status.LastChange.Equal(when) || status.LastChange.After(changedTime)
					})
				})
			})
		})
	})

	t.Run("Given an existing directory", func(t *testing.T) {
		fileName := "/tmp/exiting-directory"
		mockFS.GivenDirectoryTree(fileName)

		t.Run("When asked to watch the directory recursively", func(t *testing.T) {
			ctx, span := tracing.Start(topCtx, t.Name())
			t.Cleanup(func() {
				if t.Failed() {
					span.SetStatus(codes.Error, "test failure")
				}
				span.End()
			})

			ref := resources.RequireCreate(t, ctx, plaid.Store, AlphaV1Type, refNames[2], AlphaV1Spec{AbsolutePath: fileName, Recursive: true})

			t.Run("Then no initial change is recorded", func(t *testing.T) {
				AssertNoLastChange(t, ctx, plaid, ref)
			})

			t.Run("A a file is changed under the directory", func(t *testing.T) {
				changedPoint := time.Now()
				mockFS.GivenDirectoryTree(fileName + "/some/file")
				mockFS.FileChanged(t, ctx, fileName+"/some/file/changed")

				t.Run("Then the file records the change time", func(t *testing.T) {
					resources.AssertStatus(t, ctx, plaid.Store, ref, func(status AlphaV1Status) bool {
						return status.LastChange != nil && (changedPoint.Before(*status.LastChange) || changedPoint.Equal(*status.LastChange))
					})
				})
			})
		})
	})
}

func AssertNoLastChange(t *testing.T, ctx context.Context, plaid *resources.TestSubsystem, ref resources.Meta) {
	resources.AssertStatus(t, ctx, plaid.Store, ref, func(status AlphaV1Status) bool {
		return status.LastChange == nil
	})
}

func traceSubtest(t *testing.T, parentContext context.Context) (context.Context, trace.Span) {
	ctx, span := tracing.Start(parentContext, t.Name())
	t.Cleanup(func() {
		if t.Failed() {
			span.SetStatus(codes.Error, "test failed")
		}
		span.End()
	})
	return ctx, span
}
