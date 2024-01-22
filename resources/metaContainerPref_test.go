package resources

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"math/rand"
	"testing"
)

// 2023-03-09 @ 10:04a -- Baseline with outer type composite key + find then upsert when missing
// BenchmarkMetaContainer-8         2569396               477.1 ns/op           421 B/op          4 allocs/op
func BenchmarkMetaContainer(b *testing.B) {
	refs := make([]Meta, b.N)
	for i := 0; i < b.N; i++ {
		suffix := rand.Int31n(int32(i + 1))
		nodeRef := Meta{
			Type: Type{
				Kind:    faker.DomainName(),
				Version: faker.Word(),
			},
			Name: faker.Word() + "-" + fmt.Sprintf("%d", suffix),
		}
		refs[i] = nodeRef
	}
	c := NewMetaContainer[node]()
	b.ResetTimer()

	b.Run("Upsert", func(b *testing.B) {
		for _, nodeRef := range refs {
			_, has := c.Find(nodeRef)
			if has {
				continue
			}
			c.Upsert(nodeRef, &node{})
		}
	})
}
