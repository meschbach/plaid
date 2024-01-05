package fileWatch

import (
	"context"
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"path/filepath"
	"testing"
)

func NewMockFileWatcher(plaid *resources.TestSubsystem) *MockFS {
	mock := &MockFS{
		changeEvent: make(chan changeEvent, 16),
	}
	plaid.AttachController("file-watcher-controller", NewController(plaid.Controller, mock))
	return mock
}

type MockFS struct {
	changeEvent      chan changeEvent
	watching         []string
	knownDirectories []string
	knownFiles       []string
}

func (m *MockFS) ListDirectories(path string) ([]string, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("only absolute file paths accepted, got %s", path)
	}
	var dirs []string
	for _, d := range m.knownDirectories {
		parentDir := filepath.Dir(d)
		if parentDir == path {
			dirs = append(dirs, d)
		}
	}
	return dirs, nil
}

func (m *MockFS) Watch(ctx context.Context, path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("only absolute file paths accepted, got %s", path)
	}
	m.watching = append(m.watching, path)
	return nil
}

func (m *MockFS) ChangeFeed() <-chan changeEvent {
	return m.changeEvent
}

func (m *MockFS) GivenFileExists(fileName string) {
	m.knownDirectories = append(m.knownDirectories, filepath.Dir(fileName))
	m.knownFiles = append(m.knownFiles, fileName)
}

func (m *MockFS) GivenDirectoryTree(dirName string) {
	m.knownFiles = append(m.knownFiles, dirName)
}

func (m *MockFS) FileChanged(t *testing.T, parent context.Context, fileName string) {
	ctx, span := tracing.Start(parent, "MockFS.FileChanged", trace.WithAttributes(attribute.String("file.name", fileName)))
	defer span.End()

	dir := filepath.Dir(fileName)
	found := false
	for _, d := range m.knownFiles {
		if d == dir {
			found = true
			break
		}
	}
	if !found {
		for _, d := range m.knownDirectories {
			if d == dir {
				found = true
				break
			}
		}
	}
	if !found {
		panic(fmt.Sprintf("no such directory %q in %#v", dir, m.knownFiles))
	}

	select {
	case m.changeEvent <- changeEvent{
		path:    fileName,
		kind:    fsFileModified,
		tracing: trace.SpanContextFromContext(ctx),
	}:
	case <-ctx.Done():
		span.SetStatus(codes.Error, "context done")
		span.RecordError(ctx.Err())
		assert.NoError(t, ctx.Err())
	}
}
