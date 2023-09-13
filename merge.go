package ard

// Deep merge of source value into target value. [Map] and [StringMap]
// are merged key by key, recursively.
//
// When appendList is true then target list elements are appended to the
// source list, otherwise the source list is overridden (copied over).
//
// The source value remains safe, in that all merged data is copied (via
// [Copy]) into the target, thus any changes made to the target will not
// affect the source. On the other hand, the target value is changed
// in place. Note that for arguments that are not [Map] or [StringMap]
// you must use the return value from this function because it may return
// a new target value, e.g. a new [List] slice when appendList is true.
// Thus a safe way to use this function is like so:
//
// target = Merge(target, source, true)
func Merge(target Value, source Value, appendLists bool) Value {
	if targetMap, ok := target.(Map); ok {
		if sourceMap, ok := source.(Map); ok {
			for key, sourceValue := range sourceMap {
				if targetValue, ok := targetMap[key]; ok {
					// Target key already exists, so merge
					targetMap[key] = Merge(targetValue, sourceValue, appendLists)
				} else {
					// Target key doesn't exist, so copy
					targetMap[Copy(key)] = Copy(sourceValue)
				}
			}

			return targetMap
		}
	}

	if targetMap, ok := target.(StringMap); ok {
		if sourceMap, ok := source.(StringMap); ok {
			for key, sourceValue := range sourceMap {
				if targetValue, ok := targetMap[key]; ok {
					// Target key already exists, so merge
					targetMap[key] = Merge(targetValue, sourceValue, appendLists)
				} else {
					// Target key doesn't exist, so copy
					targetMap[key] = Copy(sourceValue)
				}
			}

			return targetMap
		}
	}

	if appendLists {
		if targetList, ok := target.(List); ok {
			if sourceList, ok := source.(List); ok {
				for _, sourceValue := range sourceList {
					targetList = append(targetList, Copy(sourceValue))
				}
				return targetList
			}
		}
	}

	return Copy(source)
}
