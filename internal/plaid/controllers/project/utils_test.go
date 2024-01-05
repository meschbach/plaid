package project

import (
	"context"
	"errors"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/resources"
	"time"
)

type BuildRunStatusMock struct {
	Ref    *resources.Meta
	Store  *resources.Client
	status buildrun.AlphaStatus1
}

func (m *BuildRunStatusMock) FinishNow(ctx context.Context, exitCode int) error {
	m.status.Ready = true
	now := time.Now()
	m.status.Build.Result = &exec.InvocationAlphaV1Status{
		Started:    &now,
		Finished:   &now,
		ExitStatus: &exitCode,
		Healthy:    true,
	}
	m.status.Run.Result = &exec.InvocationAlphaV1Status{
		Started:    &now,
		Finished:   &now,
		ExitStatus: &exitCode,
		Healthy:    true,
	}
	m.status.Ready = true
	return m.update(ctx)
}

func (m *BuildRunStatusMock) update(ctx context.Context) error {
	if m.Ref == nil {
		return errors.New("ref is nil")
	}
	exists, err := m.Store.UpdateStatus(ctx, *m.Ref, m.status)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("does not exist")
	}
	return nil
}
