package project

import (
	"context"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/optest"
	"testing"
	"time"
)

func MustFinishBuildRunRun(t *testing.T, ctx context.Context, store *resources.Client, buildRunRef resources.Meta, exitCode int) {
	ready := exitCode >= 0
	now := time.Now()
	status := buildrun.AlphaStatus1{
		Run: buildrun.Alpha1StatusRun{
			Result: &exec.InvocationAlphaV1Status{
				Started:    &now,
				Finished:   &now,
				ExitStatus: &exitCode,
				Healthy:    true,
			},
		},
		Ready: ready,
	}
	optest.MustUpdateStatusRaw(t, ctx, store, buildRunRef, status)
}
