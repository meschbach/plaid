package jtest

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func contextFromEnvTimeout(t *testing.T) (context.Context, func()) {
	base := context.Background()

	timeoutText, has := os.LookupEnv("TEST_TIMEOUT")
	if !has {
		return context.WithCancel(base)
	} else {
		timeout, err := time.ParseDuration(timeoutText)
		require.NoError(t, err, "bad test timeout given: %s", timeoutText)
		return context.WithTimeout(base, timeout)
	}
}

func ContextFromEnv(t *testing.T) context.Context {
	ctx, done := contextFromEnvTimeout(t)
	t.Cleanup(done)
	return ctx
}
