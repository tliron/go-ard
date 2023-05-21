package ard

// Ensure data adheres to the ARD map type
// (JSON decoding uses map[string]any instead of map[any]any)
func NormalizeMaps(value Value) (Value, bool) {
	switch value_ := value.(type) {
	case StringMap:
		return StringMapToMap(value_), true

	case Map:
		changedMap := make(Map)
		changed := false
		for key, element := range value_ {
			var changedKey bool
			key, changedKey = NormalizeMaps(key)

			var changedElement bool
			element, changedElement = NormalizeMaps(element)

			if changedKey || changedElement {
				changed = true
			}

			changedMap[key] = element
		}
		if changed {
			return changedMap, true
		}

	case List:
		changedList := make(List, len(value_))
		changed := false
		for index, element := range value_ {
			var changed_ bool
			if element, changed_ = NormalizeMaps(element); changed_ {
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

func NormalizeStringMaps(value Value) (Value, bool) {
	switch value_ := value.(type) {
	case Map:
		return MapToStringMap(value_), true

	case StringMap:
		changedStringMap := make(StringMap)
		changed := false
		for key, element := range value_ {
			var changed_ bool
			if element, changed_ = NormalizeStringMaps(element); changed_ {
				changed = true
			}
			changedStringMap[key] = element
		}
		if changed {
			return changedStringMap, true
		}

	case List:
		changedList := make(List, len(value_))
		changed := false
		for index, element := range value_ {
			var changed_ bool
			if element, changed_ = NormalizeStringMaps(element); changed_ {
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
