package resources

import (
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTypeStruct(t *testing.T) {
	t.Run("Type Validity", func(t *testing.T) {
		t.Run("Default Type", func(t *testing.T) {
			var exampleType Type
			assert.False(t, exampleType.Valid())
		})

		t.Run("Type with kind but not version", func(t *testing.T) {
			assert.False(t, Type{Kind: faker.Word()}.Valid())
		})

		t.Run("Type with version but not kind", func(t *testing.T) {
			assert.False(t, Type{Version: faker.Word()}.Valid())
		})

		t.Run("Type with values is valid", func(t *testing.T) {
			exampleType := FakeType()
			assert.True(t, exampleType.Valid())
		})
	})
}
