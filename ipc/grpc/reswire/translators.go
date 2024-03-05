package reswire

import (
	"fmt"
	"github.com/meschbach/plaid/resources"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ExternalizeType(t resources.Type) *Type {
	return &Type{
		Kind:    t.Kind,
		Version: t.Version,
	}
}

func MetaToWire(ref resources.Meta) *Meta {
	return &Meta{
		Kind: ExternalizeType(ref.Type),
		Name: ref.Name,
	}
}

func InternalizeKind(in *Type) resources.Type {
	return resources.Type{
		Kind:    in.Kind,
		Version: in.Version,
	}
}

func InternalizeMeta(in *Meta) resources.Meta {
	return resources.Meta{
		Type: InternalizeKind(in.Kind),
		Name: in.Name,
	}
}

func ExternalizeEventLevel(l resources.EventLevel) EventLevel {
	switch l {
	case resources.AllEvents:
		return EventLevel_All
	case resources.Info:
		return EventLevel_Info
	case resources.Error:
		return EventLevel_Error
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func InternalizeEventLevel(l EventLevel) resources.EventLevel {
	switch l {
	case EventLevel_All:
		return resources.AllEvents
	case EventLevel_Error:
		return resources.Error
	case EventLevel_Info:
		return resources.Info
	default:
		panic(fmt.Sprintf("unhandled translation from %d", l))
	}
}

func Eventf(when time.Time, level resources.EventLevel, format string, args ...any) *Event {
	message := fmt.Sprintf(format, args...)
	wireLevel := ExternalizeEventLevel(level)
	wireWhen := timestamppb.New(when)
	return &Event{
		When:     wireWhen,
		Level:    wireLevel,
		Rendered: message,
	}
}

func InternalizeOperation(op WatcherEventOut_Op) resources.ResourceChangedOperation {
	switch op {
	case WatcherEventOut_Created:
		return resources.CreatedEvent
	case WatcherEventOut_UpdatedStatus:
		return resources.StatusUpdated
	case WatcherEventOut_Deleted:
		return resources.DeletedEvent
	default:
		panic(fmt.Sprintf("unknown value %q", op.String()))
	}
}
