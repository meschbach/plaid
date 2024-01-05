package resources

import "strings"

// MetaContainer is a container which will associate a set of Meta with some value.  Designed to be optimized for the
// underlying meta structure.
type MetaContainer[T any] struct {
	//buckets contains kind/version -> name -> object type
	//experiment note from 2023-03-09: tried splitting the 'kind/version' map into two maps with different levels.
	//this resulted in the following changes.  My guess is the map[string]... is more expensive than the string concat
	//BenchmarkMetaContainer-8                         2181157               537.3 ns/op           438 B/op          4 allocs/op
	//BenchmarkMetaContainer_SeparateBucket-8          2159016               627.0 ns/op           674 B/op          5 allocs/op
	buckets map[string]map[string]*T
}

func NewMetaContainer[T any]() *MetaContainer[T] {
	return &MetaContainer[T]{}
}

func (m *MetaContainer[T]) typeKeyOf(which Type) string {
	return which.Kind + "/" + which.Version
}
func (m *MetaContainer[T]) typeKey(which Meta) string {
	return m.typeKeyOf(which.Type)
}

func (m *MetaContainer[T]) GetOrCreate(which Meta, onCreate func() *T) (value *T, created bool) {
	created = false
	if m.buckets == nil {
		m.buckets = make(map[string]map[string]*T)
		created = true
	}

	k := m.typeKey(which)
	kindBucket, hasType := m.buckets[k]
	if !hasType {
		created = true
		kindBucket = make(map[string]*T)
		m.buckets[k] = kindBucket
	}
	current, hasOld := kindBucket[which.Name]
	if !hasOld {
		created = true
		current = onCreate()
		kindBucket[which.Name] = current
	}
	return current, created
}

// Find locates the value stored for which
func (m *MetaContainer[T]) Find(which Meta) (value *T, found bool) {
	if m.buckets == nil {
		return nil, false
	}
	k := m.typeKey(which)
	kindBucket, hasType := m.buckets[k]
	if !hasType {
		return nil, false
	}
	t, has := kindBucket[which.Name]
	if !has {
		return nil, false
	}
	return t, true
}

func (m *MetaContainer[T]) Upsert(which Meta, value *T) (*T, bool) {
	if m.buckets == nil {
		m.buckets = make(map[string]map[string]*T)
	}

	k := m.typeKey(which)
	kindBucket, hasType := m.buckets[k]
	if !hasType {
		kindBucket = make(map[string]*T)
		m.buckets[k] = kindBucket
	}
	old, hasOld := kindBucket[which.Name]
	kindBucket[which.Name] = value
	return old, hasOld
}

func (m *MetaContainer[T]) AllValues() []*T {
	var values []*T
	for _, t := range m.buckets {
		for _, v := range t {
			values = append(values, v)
		}
	}
	return values
}

func (m *MetaContainer[T]) Delete(key Meta) (*T, bool) {
	if m.buckets == nil {
		return nil, false
	}
	k := m.typeKey(key)
	kindBucket, hasType := m.buckets[k]
	if !hasType {
		return nil, false
	}
	t, has := kindBucket[key.Name]
	if !has {
		return nil, false
	}
	return t, true
}

func (m *MetaContainer[T]) ListNames(forType Type) []string {
	if m.buckets == nil {
		return nil
	}
	typeKey := m.typeKeyOf(forType)
	typeResources, hasType := m.buckets[typeKey]
	if !hasType {
		return nil
	}
	out := make([]string, 0, len(typeResources))
	for name := range typeResources {
		out = append(out, name)
	}
	return out
}

func (m *MetaContainer[T]) AllTypes() []Type {
	var output []Type
	for encodedType := range m.buckets {
		parts := strings.Split(encodedType, "/")
		output = append(output, Type{
			Kind:    parts[0],
			Version: parts[1],
		})
	}
	return output
}
