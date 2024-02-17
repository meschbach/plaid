package resources

type metaContainerNode[T any] struct {
	values map[string]*T
}

// MetaContainer is a container which will associate a set of Meta with some value.  Designed to be optimized for the
// underlying meta structure.
type MetaContainer[T any] struct {
	//buckets contains kind/version -> name -> object type
	//experiment note from 2023-03-09: tried splitting the 'kind/version' map into two maps with different levels.
	//this resulted in the following changes.  My guess is the map[string]... is more expensive than the string concat
	//BenchmarkMetaContainer-8                         2181157               537.3 ns/op           438 B/op          4 allocs/op
	//BenchmarkMetaContainer_SeparateBucket-8          2159016               627.0 ns/op           674 B/op          5 allocs/op
	buckets *TypeContainer[metaContainerNode[T]]
}

func NewMetaContainer[T any]() *MetaContainer[T] {
	return &MetaContainer[T]{
		buckets: NewTypeContainer[metaContainerNode[T]](),
	}
}

//
//func (m *MetaContainer[T]) typeKeyOf(which Type) string {
//	return which.Kind + "/" + which.Version
//}
//func (m *MetaContainer[T]) typeKey(which Meta) string {
//	return m.typeKeyOf(which.Type)
//}

func (m *MetaContainer[T]) GetOrCreate(which Meta, onCreate func() *T) (value *T, created bool) {
	created = false
	if m.buckets == nil {
		m.buckets = NewTypeContainer[metaContainerNode[T]]()
	}

	node, _ := m.buckets.GetOrCreate(which.Type, func() *metaContainerNode[T] {
		return &metaContainerNode[T]{
			values: make(map[string]*T),
		}
	})
	value, hasOld := node.values[which.Name]
	if !hasOld {
		created = true
		value = onCreate()
		node.values[which.Name] = value
	}
	return value, created
}

// Find locates the value stored for which
func (m *MetaContainer[T]) Find(which Meta) (value *T, found bool) {
	if m.buckets == nil {
		return nil, false
	}
	bucket, hasBucket := m.buckets.Find(which.Type)
	if !hasBucket {
		return nil, false
	}
	t, has := bucket.values[which.Name]
	return t, has
}

func (m *MetaContainer[T]) Upsert(which Meta, value *T) (*T, bool) {
	if m.buckets == nil {
		m.buckets = NewTypeContainer[metaContainerNode[T]]()
	}

	node, _ := m.buckets.GetOrCreate(which.Type, func() *metaContainerNode[T] {
		return &metaContainerNode[T]{values: make(map[string]*T)}
	})
	old, hasOld := node.values[which.Name]
	node.values[which.Name] = value
	return old, hasOld
}

func (m *MetaContainer[T]) AllValues() []*T {
	var values []*T
	for _, node := range m.buckets.AllValues() {
		for _, v := range node.values {
			values = append(values, v)
		}
	}
	return values
}

func (m *MetaContainer[T]) Delete(key Meta) (*T, bool) {
	if m.buckets == nil {
		return nil, false
	}
	node, has := m.buckets.Find(key.Type)
	if !has {
		return nil, false
	}
	t, has := node.values[key.Name]
	if !has {
		return nil, false
	}
	delete(node.values, key.Name)
	return t, true
}

func (m *MetaContainer[T]) ListNames(forType Type) []string {
	if m.buckets == nil {
		return nil
	}
	node, has := m.buckets.Find(forType)
	if !has {
		return nil
	}
	out := make([]string, 0, len(node.values))
	for name := range node.values {
		out = append(out, name)
	}
	return out
}

func (m *MetaContainer[T]) AllTypes() []Type {
	if m.buckets == nil {
		return nil
	}
	return m.buckets.AllTypes()
}

// AllMetas will create an comprehensive list of all Meta names stored within the structure.  Result orders are
// nondeterministic.
func (m *MetaContainer[T]) AllMetas() []Meta {
	var found []Meta
	types := m.AllTypes()
	for _, t := range types {
		bucket, _ := m.buckets.Find(t)
		for name, _ := range bucket.values {
			found = append(found, Meta{
				Type: t,
				Name: name,
			})
		}
	}
	return found
}
