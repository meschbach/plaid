package resources

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"math/rand"
	"testing"
)

// 2023-03-09 @ 9:59a -- Initial baseline
// BenchmarkControllerGetOrCreateNode-8     2663437               495.5 ns/op           418 B/op          4 allocs/op
// 1.20.2 - Reusing MetaContainer
// BenchmarkControllerGetOrCreateNode-8     2767586               460.6 ns/op           415 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2486145               443.6 ns/op           424 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2645773               468.1 ns/op           418 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2856108               422.4 ns/op           413 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2605802               431.4 ns/op           420 B/op          4 allocs/op
func BenchmarkControllerGetOrCreateNode(b *testing.B) {
	refs := make([]Meta, b.N)
	for i := 0; i < b.N; i++ {
		suffix := rand.Int31n(int32(i + 1))
		nodeRef := Meta{
			Type: Type{
				Kind:    faker.DomainName(),
				Version: faker.Word(),
			},
			Name: "some-name-" + fmt.Sprintf("%d", suffix),
		}
		refs[i] = nodeRef
	}
	c := NewController()
	b.ResetTimer()

	for _, nodeRef := range refs {
		c.createOrGetNode(nil, nodeRef)
	}
}
