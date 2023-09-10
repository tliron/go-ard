package ard

// Deep copy.
//
// The input can be a mix of ARD and non-ARD values (e.g. Go structs). Non-ARD
// values will be left as is, thus the returned may not be valid ARD. To ensure
// a valid ARD result use [ValidCopy].
//
// Recurses into [Map], [StringMap], and [List], creating new instances of
// each. Thus a [Map] is copied into a new [Map] and a [StringMap] is copied
// into a new [StringMap]. To convert them to a unified map type use
// [CopyStringMapsToMaps] or [CopyMapsToStringMaps].
func Copy(value Value) Value {
	value, _ = copy_(value, nil, noConversion)
	return value
}

// Like [Copy] but converts all [StringMap] to [Map].
//
// For in-place conversion use [ConvertStringMapsToMaps].
func CopyStringMapsToMaps(value Value) Value {
	value, _ = copy_(value, nil, convertStringMapsToMaps)
	return value
}

// Like [Copy] but converts all [Map] to [StringMap].
//
// Keys are converted using [MapKeyToString].
//
// For in-place conversion use [ConvertStringMapsToMaps].
func CopyMapsToStringMaps(value Value) Value {
	value, _ = copy_(value, nil, convertMapsToStringMaps)
	return value
}

// Deep copy and return a valid ARD value.
//
// The input can be a mix of ARD and non-ARD values (e.g. Go structs). The
// returned value is guaranteed to be valid ARD. This works by reflecting any non-ARD
// via the provided [*Reflector]. The reflector argument can be nil, in which case a
// default reflector will be used. To leave non-ARD values as is use [Copy].
//
// This function can be used to "canonicalize" values to ARD, for which is should
// generally be more efficient than calling [Roundtrip].
//
// Recurses into [Map], [StringMap], and [List], creating new instances of
// each. Thus a [Map] is copied into a new [Map] and a [StringMap] is copied
// into a new [StringMap]. To convert them to a unified map type use
// [ValidCopyStringMapsToMaps] or [ValidCopyMapsToStringMaps].
func ValidCopy(value Value, reflector *Reflector) (Value, error) {
	if reflector == nil {
		reflector = NewReflector()
	}

	return copy_(value, reflector, noConversion)
}

// Like [ValidCopy] but converts all [StringMap] to [Map].
//
// For in-place conversion use [ConvertStringMapsToMaps].
func ValidCopyStringMapsToMaps(value Value, reflector *Reflector) (Value, error) {
	if reflector == nil {
		reflector = NewReflector()
	}

	return copy_(value, reflector, convertStringMapsToMaps)
}

// Like [ValidCopy] but converts all [Map] to [StringMap].
//
// Keys are converted using [MapKeyToString].
//
// For in-place conversion use [ConvertMapsToStringMaps].
func ValidCopyMapsToStringMaps(value Value, reflector *Reflector) (Value, error) {
	if reflector == nil {
		reflector = NewReflector()
	}

	return copy_(value, reflector, convertMapsToStringMaps)
}

// When reflector is nil will never return an error.
func copy_(value Value, reflector *Reflector, mode conversionMode) (Value, error) {
	var err error
	switch value_ := value.(type) {
	case Map:
		if mode == convertMapsToStringMaps {
			copiedMap := make(StringMap)
			for key, value__ := range value_ {
				if copiedMap[MapKeyToString(key)], err = copy_(value__, reflector, mode); err != nil {
					return nil, err
				}
			}
			return copiedMap, nil
		} else {
			copiedMap := make(Map)
			for key, value__ := range value_ {
				if copiedMap[key], err = copy_(value__, reflector, mode); err != nil {
					return nil, err
				}
			}
			return copiedMap, nil
		}

	case StringMap:
		if mode == convertStringMapsToMaps {
			copiedMap := make(Map)
			for key, value__ := range value_ {
				if copiedMap[key], err = copy_(value__, reflector, mode); err != nil {
					return nil, err
				}
			}
			return copiedMap, nil
		} else {
			copiedMap := make(StringMap)
			for key, value__ := range value_ {
				if copiedMap[key], err = copy_(value__, reflector, mode); err != nil {
					return nil, err
				}
			}
			return copiedMap, nil
		}

	case List:
		copiedList := make(List, len(value_))
		for index, entry := range value_ {
			if copiedList[index], err = copy_(entry, reflector, mode); err != nil {
				return nil, err
			}
		}
		return copiedList, nil

	default:
		if reflector != nil {
			if IsPrimitiveType(value) {
				return value, nil
			} else {
				if value, err = reflector.Unpack(value, mode == convertMapsToStringMaps); err == nil {
					if mode != noConversion {
						value, _ = convert(value, mode)
					}
					return value, nil
				} else {
					return nil, err
				}
			}
		} else {
			return value, nil
		}
	}
}
