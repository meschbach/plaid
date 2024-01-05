package projectfile

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/files"
	"path/filepath"
	"time"
)

type fileState struct {
	hasBeenParsed bool
	lastParsed    time.Time
	parsedFile    Configuration
	lastError     *string
}

type fileStateNext int

const (
	fileStateParse fileStateNext = iota
	fileStateDone
)

func (f *fileState) updateStatus(spec Alpha1Spec, status *Alpha1Status) {
	status.ParsingError = f.lastError
	status.LastLoaded = &f.lastParsed
	status.ProjectFile = filepath.Join(spec.WorkingDirectory, spec.ProjectFile)
}

func (f *fileState) decideNextStep(ctx context.Context) (fileStateNext, error) {
	if !f.hasBeenParsed {
		return fileStateParse, nil
	}
	return fileStateDone, nil
}

func (f *fileState) parse(ctx context.Context, spec Alpha1Spec) error {
	f.hasBeenParsed = true
	projectFileName := filepath.Join(spec.WorkingDirectory, spec.ProjectFile)
	if err := files.ParseJSONFile(projectFileName, &f.parsedFile); err != nil {
		msg := err.Error()
		f.lastError = &msg
		return err
	}
	f.lastParsed = time.Now()
	return nil
}
