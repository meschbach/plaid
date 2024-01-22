package resources

import (
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTypeContainer(t *testing.T) {
	t.Run("Given an empty type container", func(t *testing.T) {
		container := NewTypeContainer[int]()

		t.Run("When inserting a new type", func(t *testing.T) {
			exampleType := Type{
				Kind:    faker.Word(),
				Version: faker.Word(),
			}
			value := 42
			old, hasOld := container.Upsert(exampleType, &value)
			if assert.False(t, hasOld, "then there is no old value") {
				assert.Nil(t, old, "then the old value is nil")
			}

			t.Run("And when retrieved", func(t *testing.T) {
				foundValue, has := container.Find(exampleType)
				if assert.True(t, has, "Then the container has the value") {
					assert.Equal(t, value, *foundValue, "Then the found value is equal to the original")
				}
			})

			t.Run("And asked to find or create existing value", func(t *testing.T) {
				subjectValue, has := container.GetOrCreate(exampleType, func() *int {
					return nil
				})
				if assert.False(t, has, "Then the container has a value") {
					assert.NotNil(t, value, *subjectValue, "Then the value is not nil")
				}
			})

			t.Run("And when queried for all values", func(t *testing.T) {
				values := container.AllValues()
				if assert.Len(t, values, 1, "Then there is a single value") {
					assert.Equal(t, value, *values[0], "Then the value is equal to the original")
				}
			})

			t.Run("And when all types are queried", func(t *testing.T) {
				types := container.AllTypes()
				if assert.Len(t, types, 1, "Then there is a single type") {
					assert.Equal(t, exampleType, types[0], "Then the sample type is returned")
				}
			})

			t.Run("And a another value is inserted", func(t *testing.T) {
				newValue := 46
				old, hasOld := container.Upsert(exampleType, &newValue)
				if assert.True(t, hasOld, "then there is an old value") {
					assert.Equal(t, value, *old, "then the old value is equal to the original value")
				}
			})
		})
	})
}
