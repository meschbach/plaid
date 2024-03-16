package resources

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMeta(t *testing.T) {
	t.Run("Given a meta object", func(t *testing.T) {
		t.Run("When compared to itself it is equal", func(t *testing.T) {
			ref := FakeMeta()
			assert.Truef(t, ref.EqualsMeta(ref), "identity equality")
		})
	})

	t.Run("Given a slice of meta objects", func(t *testing.T) {
		example := []Meta{FakeMeta(), FakeMeta(), FakeMeta()}

		t.Run("When an asked if an existing element is contained within then it is true", func(t *testing.T) {
			assert.True(t, MetaSliceContains(example, example[1]), "meta slice contains an element within")
		})
	})

	t.Run("Meta Validity", func(t *testing.T) {
		t.Run("Given a default meat", func(t *testing.T) {
			var meta Meta
			assert.False(t, meta.Valid())
		})

		t.Run("Given a fake type without a name", func(t *testing.T) {
			ref := FakeMetaOf(Type{})
			assert.False(t, ref.Valid())
		})

		t.Run("Given a valid ref", func(t *testing.T) {
			ref := FakeMeta()
			assert.True(t, ref.Valid())
		})
	})
}
