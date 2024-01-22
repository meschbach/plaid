package resources

import (
	"github.com/go-faker/faker/v4"
	"testing"
)

type benchmarkExampleNode struct {
	value int
}

var globalDeoptimization *benchmarkExampleNode
var globalHas bool

func BenchmarkTypeContainerInsert(b *testing.B) {
	refs := make([]Type, b.N)
	for i := 0; i < b.N; i++ {
		nodeRef := Type{
			Kind:    faker.DomainName(),
			Version: faker.Word(),
		}
		refs[0] = nodeRef
	}
	c := NewTypeContainer[benchmarkExampleNode]()
	b.ResetTimer()

	b.Run("Upsert", func(b *testing.B) {
		for i, nodeRef := range refs {
			c.Upsert(nodeRef, &benchmarkExampleNode{value: i})
		}
	})

	b.Run("Retrieval", func(b *testing.B) {
		for _, nodeRef := range refs {
			globalDeoptimization, globalHas = c.Find(nodeRef)
			if !globalHas {
				panic("must have")
			}
			if globalDeoptimization == nil {
				panic("no value")
			}
		}
	})
}
