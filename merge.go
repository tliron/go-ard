package ard

import (
	"errors"
	"strings"
)

func MergeMaps(target Map, source Map, mergeLists bool) {
	for key, sourceValue := range source {
		if targetValue, ok := target[key]; ok {
			switch sourceValue_ := sourceValue.(type) {
			case Map:
				if targetValueMap, ok := targetValue.(Map); ok {
					MergeMaps(targetValueMap, sourceValue_, mergeLists)
					continue
				}

			case List:
				if mergeLists {
					if targetValueList, ok := targetValue.(List); ok {
						target[key] = append(targetValueList, sourceValue_...)
						continue
					}
				}
			}
		}

		target[key] = Copy(sourceValue)
	}
}

func MergeStringMaps(target StringMap, source StringMap, mergeLists bool) {
	for key, sourceValue := range source {
		if targetValue, ok := target[key]; ok {
			switch sourceValue_ := sourceValue.(type) {
			case StringMap:
				if targetValueMap, ok := targetValue.(StringMap); ok {
					MergeStringMaps(targetValueMap, sourceValue_, mergeLists)
					continue
				}

			case List:
				if mergeLists {
					if targetValueList, ok := targetValue.(List); ok {
						target[key] = append(targetValueList, sourceValue_...)
						continue
					}
				}
			}
		}

		target[key] = Copy(sourceValue)
	}
}

// TODO: use Node instead
func StringMapPutNested(map_ StringMap, key string, value string) error {
	path := strings.Split(key, ".")
	last := len(path) - 1

	if last == -1 {
		return errors.New("empty key")
	}

	if last > 0 {
		for _, p := range path[:last] {
			if o, ok := map_[p]; ok {
				if map_, ok = o.(StringMap); !ok {
					return errors.New("bad nested map structure")
				}
			} else {
				m := make(StringMap)
				map_[p] = m
				map_ = m
			}
		}
	}

	map_[path[last]] = value

	return nil
}
