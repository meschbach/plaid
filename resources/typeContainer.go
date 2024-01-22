package resources

type TypeContainer[T any] struct {
	values map[string]map[string]*T
}

func NewTypeContainer[T any]() *TypeContainer[T] {
	return &TypeContainer[T]{
		values: make(map[string]map[string]*T),
	}
}

// GetOrCreate will return the existing node if it exists for a type, otherwise will invoke onCreate to create the value
// for the node. Will return the node (new or existing) and a boolean to indicate teh value has been created.
func (t *TypeContainer[T]) GetOrCreate(which Type, onCreate func() *T) (node *T, created bool) {
	created = false
	kindNode, hasKind := t.values[which.Kind]
	if !hasKind {
		kindNode = make(map[string]*T)
		t.values[which.Kind] = kindNode
	}
	versionNode, hasVersion := kindNode[which.Version]
	if !hasVersion {
		created = true
		versionNode = onCreate()
		kindNode[which.Version] = versionNode
	}
	return versionNode, created
}

func (t *TypeContainer[T]) Find(which Type) (*T, bool) {
	kindNode, hasKind := t.values[which.Kind]
	if !hasKind {
		return nil, false
	}
	versionNode, hasVersion := kindNode[which.Version]
	if !hasVersion {
		return nil, false
	}
	return versionNode, true
}

func (t *TypeContainer[T]) Upsert(which Type, value *T) (*T, bool) {
	hasOld := true
	kindNode, hasKind := t.values[which.Kind]
	if !hasKind {
		hasOld = false
		kindNode = make(map[string]*T)
		t.values[which.Kind] = kindNode
	}
	old, hasVersion := kindNode[which.Version]
	if !hasVersion {
		hasOld = false
	}
	kindNode[which.Version] = value
	return old, hasOld
}

func (t *TypeContainer[T]) AllValues() []*T {
	out := make([]*T, 0, len(t.values))
	for _, kindNode := range t.values {
		for _, versionNode := range kindNode {
			out = append(out, versionNode)
		}
	}
	return out
}

func (t *TypeContainer[T]) AllTypes() []Type {
	out := make([]Type, 0, len(t.values))
	for kindName, kindNode := range t.values {
		for versionName, _ := range kindNode {
			out = append(out, Type{
				Kind:    kindName,
				Version: versionName,
			})
		}
	}
	return out
}
