package ard

// Converts any [StringMap] to [Map] recursively, ensuring that no
// [StringMap] will be present. Conversion happens in place, unless the
// input is itself a [StringMap], in which case a new [Map] will be
// returned.
//
// Returns true if any conversion occurred.
func ConvertStringMapsToMaps(value Value) (Value, bool) {
	return convert(value, convertStringMapsToMaps)
}

// Converts any [Map] to [StringMap] recursively, ensuring that no
// [Map] will be present. Conversion happens in place, unless the
// input is itself a [Map], in which case a new [StringMap] will be
// returned.
//
// Keys are converted using [MapKeyToString].
//
// Returns true if any conversion occurred.
func ConvertMapsToStringMaps(value Value) (Value, bool) {
	return convert(value, convertMapsToStringMaps)
}

func convert(value Value, mode conversionMode) (Value, bool) {
	switch value_ := value.(type) {
	case StringMap:
		if mode == convertStringMapsToMaps {
			changedMap := make(Map)

			for key, value__ := range value_ {
				changedMap[key], _ = convert(value__, mode)
			}

			return changedMap, true
		} else {
			changedStringMap := make(StringMap)
			changed := false

			for key, element := range value_ {
				var changed_ bool
				if element, changed_ = convert(element, mode); changed_ {
					changed = true
				}
				changedStringMap[key] = element
			}

			if changed {
				return changedStringMap, true
			}
		}

	case Map:
		if mode == convertMapsToStringMaps {
			changedMap := make(StringMap)

			for key, value__ := range value_ {
				changedMap[MapKeyToString(key)], _ = convert(value__, mode)
			}

			return changedMap, true
		} else {
			changedMap := make(Map)
			changed := false

			for key, element := range value_ {
				var changedKey bool
				key, changedKey = convert(key, mode)

				var changedElement bool
				element, changedElement = convert(element, mode)

				if changedKey || changedElement {
					changed = true
				}

				changedMap[key] = element
			}

			if changed {
				return changedMap, true
			}
		}

	case List:
		changedList := make(List, len(value_))
		changed := false

		for index, element := range value_ {
			var changed_ bool
			if element, changed_ = convert(element, mode); changed_ {
				changed = true
			}

			changedList[index] = element
		}

		if changed {
			return changedList, true
		}
	}

	return value, false
}
