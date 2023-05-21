package ard

func CopyToARD(value Value) (Value, error) {
	if IsPrimitiveType(value) {
		return value, nil
	} else {
		var err error
		switch value_ := value.(type) {
		case Map:
			map_ := make(Map)
			for key, value_ := range value_ {
				if map_[key], err = CopyToARD(value_); err != nil {
					return nil, err
				}
			}
			return map_, nil

		case StringMap:
			map_ := make(StringMap)
			for key, value_ := range value_ {
				if map_[key], err = CopyToARD(value_); err != nil {
					return nil, err
				}
			}
			return map_, nil

		case List:
			list := make(List, len(value_))
			for index, entry := range value_ {
				if list[index], err = CopyToARD(entry); err != nil {
					return nil, err
				}
			}
			return list, nil

		default:
			// TODO: not very efficient
			return Roundtrip(value, "")
		}
	}
}

func NormalizeMapsCopyToARD(value Value) (Value, error) {
	if value_, err := CopyToARD(value); err == nil {
		value_, _ = NormalizeMaps(value_)
		return value_, nil
	} else {
		return nil, err
	}
}

func NormalizeStringMapsCopyToARD(value Value) (Value, error) {
	if value_, err := CopyToARD(value); err == nil {
		value_, _ = NormalizeStringMaps(value_)
		return value_, nil
	} else {
		return nil, err
	}
}

// Will leave primitive and non-ARD types as is
func SimpleCopy(value Value) Value {
	switch value_ := value.(type) {
	case Map:
		map_ := make(Map)
		for key, value_ := range value_ {
			map_[key] = SimpleCopy(value_)
		}
		return map_

	case StringMap:
		map_ := make(StringMap)
		for key, value_ := range value_ {
			map_[key] = SimpleCopy(value_)
		}
		return map_

	case List:
		list := make(List, len(value_))
		for index, entry := range value_ {
			list[index] = SimpleCopy(entry)
		}
		return list

	default:
		return value
	}
}

func NormalizeMapsSimpleCopy(value Value) Value {
	value_ := SimpleCopy(value)
	value_, _ = NormalizeMaps(value_)
	return value_
}

func NormalizeStringMapsSimpleCopy(value Value) Value {
	value_ := SimpleCopy(value)
	value_, _ = NormalizeStringMaps(value_)
	return value_
}
