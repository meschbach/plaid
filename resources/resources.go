package resources

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Type struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

func (t Type) Equals(rhs Type) bool {
	return t.Kind == rhs.Kind && t.Version == rhs.Version
}

func (t Type) String() string {
	return fmt.Sprintf("{%s %s}", t.Kind, t.Version)
}

type Meta struct {
	Type Type   `json:"type"`
	Name string `json:"name"`
}

func (m Meta) EqualsMeta(other Meta) bool {
	return m.Type == other.Type && m.Name == other.Name
}

func (m Meta) String() string {
	return fmt.Sprintf("%s %s", m.Type, m.Name)
}

func (m Meta) AsTraceAttribute(prefix string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(prefix+".type.kind", m.Type.Kind),
		attribute.String(prefix+".type.version", m.Type.Version),
		attribute.String(prefix+".name", m.Name),
	}
}

type ResourceChangedOperation uint8

func (r ResourceChangedOperation) String() string {
	switch r {
	case CreatedEvent:
		return "created"
	case UpdatedEvent:
		return "updated"
	case DeletedEvent:
		return "deleted"
	case StatusUpdated:
		return "status-updated"
	default:
		panic("unknown status type")
	}
}

const (
	CreatedEvent ResourceChangedOperation = iota
	UpdatedEvent
	DeletedEvent
	StatusUpdated
)

type ResourceChanged struct {
	Which     Meta
	Operation ResourceChangedOperation
	Tracing   trace.Link
}

func (r ResourceChanged) String() string {
	return fmt.Sprintf("%s %s", r.Operation, r.Which)
}

func (r ResourceChanged) ToTraceAttributes() []attribute.KeyValue {
	return append(r.Which.AsTraceAttribute("change.which"), attribute.Stringer("change.operation", r.Operation))
}
