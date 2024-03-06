package service

import (
	"github.com/meschbach/plaid/ipc/grpc/reswire"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceCompliance(t *testing.T) {
	t.Run("ResourceService must fully implement gRPC service", func(t *testing.T) {
		assert.Implements(t, (*reswire.ResourceControllerServer)(nil), &ResourceService{})
	})
}
