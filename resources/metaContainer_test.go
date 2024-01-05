package resources

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetaContainer(t *testing.T) {
	t.Run("Given an empty container", func(t *testing.T) {
		c := NewMetaContainer[int]()

		t.Run("When inserting first via GetOrCreate", func(t *testing.T) {
			fakeRef := FakeMeta()
			value, created := c.GetOrCreate(fakeRef, func() *int {
				i := 1
				return &i
			})

			t.Run("Then provides new value", func(t *testing.T) {
				assert.Equal(t, 1, *value)
			})

			t.Run("Then reports resource created", func(t *testing.T) {
				assert.True(t, created, "value should register as created")
			})

			t.Run("When requesting types", func(t *testing.T) {
				types := c.AllTypes()

				t.Run("Then it contains the referenced type", func(t *testing.T) {
					if assert.Len(t, types, 1) {
						assert.Equal(t, fakeRef.Type.Kind, types[0].Kind)
						assert.Equal(t, fakeRef.Type.Version, types[0].Version)
					}
				})

				t.Run("And listing all names of the type", func(t *testing.T) {
					if assert.Len(t, types, 1) {
						return
					}
					allNames := c.ListNames(types[0])

					t.Run("Then the reference is returned", func(t *testing.T) {
						if assert.Len(t, allNames, 1) {
							assert.Equal(t, fakeRef.Name, allNames[0])
						}
					})
				})
			})
		})
	})
}
