package resources

import (
	"errors"
	"slices"
)

var InvalidMeta = errors.New("invalid meta")

func MetaSliceContains(set []Meta, which Meta) bool {
	return slices.ContainsFunc(set, func(meta Meta) bool {
		return which.EqualsMeta(meta)
	})
}

func (m Meta) Valid() bool {
	if !m.Type.Valid() {
		return false
	}
	return m.Name != ""
}

func (m Meta) ValidError() error {
	if !m.Valid() {
		return InvalidMeta
	} else {
		return nil
	}
}
