package resources

import "slices"

func MetaSliceContains(set []Meta, which Meta) bool {
	return slices.ContainsFunc(set, func(meta Meta) bool {
		return which.EqualsMeta(meta)
	})
}
