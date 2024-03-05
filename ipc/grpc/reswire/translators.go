package reswire

import (
	"fmt"
	"github.com/meschbach/plaid/internal/plaid/daemon/wire"
	"github.com/meschbach/plaid/resources"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func TypeToWire(t resources.Type) *wire.Type {
	return &wire.Type{
		Kind:    t.Kind,
		Version: t.Version,
	}
}

func MetaToWire(ref resources.Meta) *wire.Meta {
	return &wire.Meta{
		Kind: TypeToWire(ref.Type),
		Name: ref.Name,
	}
}

func ExternalizeEventLevel(l resources.EventLevel) wire.EventLevel {
	switch l {
	case resources.AllEvents:
		return wire.EventLevel_All
	case resources.Info:
		return wire.EventLevel_Info
	case resources.Error:
		return wire.EventLevel_Error
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func InternalizeEventLevel(l wire.EventLevel) resources.EventLevel {
	switch l {
	case wire.EventLevel_All:
		return resources.AllEvents
	case wire.EventLevel_Error:
		return resources.Error
	case wire.EventLevel_Info:
		return resources.Info
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func Eventf(when time.Time, level resources.EventLevel, format string, args ...any) *wire.Event {
	message := fmt.Sprintf(format, args...)
	wireLevel := ExternalizeEventLevel(level)
	wireWhen := timestamppb.New(when)
	return &wire.Event{
		When:     wireWhen,
		Level:    wireLevel,
		Rendered: message,
	}
}
