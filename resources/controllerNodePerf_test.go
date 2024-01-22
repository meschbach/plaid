package resources

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"math/rand"
	"testing"
)

var benchmarkControllerGetOrCreate *node

// 2023-03-09 @ 9:59a -- Initial baseline
// BenchmarkControllerGetOrCreateNode-8     2663437               495.5 ns/op           418 B/op          4 allocs/op
// 1.20.2 - Reusing MetaContainer
// BenchmarkControllerGetOrCreateNode-8     2767586               460.6 ns/op           415 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2486145               443.6 ns/op           424 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2645773               468.1 ns/op           418 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2856108               422.4 ns/op           413 B/op          4 allocs/op
// BenchmarkControllerGetOrCreateNode-8     2605802               431.4 ns/op           420 B/op          4 allocs/op
// 2024-01-22 @ 10:27a
// Due to optimizations this appears to be broken once I split the tests into their specific groups.  Here is the
// values under Golang 1.21.4
// BenchmarkControllerGetOrCreateNode/Upsert-8              2605802                 0.0002237 ns/op               0 B/op          0 allocs/op
// BenchmarkControllerGetOrCreateNode/Retrieval-8           2605802                 0.0002878 ns/op               0 B/op          0 allocs/op
// First one should produce allocations as it's inserting values.  Ran with the following.  Need to revisit later.
// > go test --test.benchtime 2605802x -test.bench=BenchmarkControllerGetOrCreateNode -benchmem ./resources/...
func BenchmarkControllerGetOrCreateNode(b *testing.B) {
	b.StopTimer()
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
	benchmarkController := NewController()
	b.StartTimer()

	b.Run("Upsert", func(b *testing.B) {
		for _, nodeRef := range refs {
			var err error
			benchmarkControllerGetOrCreate, _, err = benchmarkController.createOrGetNode(nil, nodeRef)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("Retrieval", func(b *testing.B) {
		for _, nodeRef := range refs {
			var err error
			benchmarkControllerGetOrCreate, _, err = benchmarkController.createOrGetNode(nil, nodeRef)
			if err != nil {
				panic(err)
			}
		}
	})
}
