package junk

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
)

func ObjAttribute[T any](name string, obj *T) attribute.KeyValue {
	return attribute.Stringer(name, objPtrStringer[T]{obj})
}

type objPtrStringer[T any] struct {
	obj *T
}

func (o objPtrStringer[T]) String() string {
	return fmt.Sprintf("%p", o.obj)
}
