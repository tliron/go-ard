package ard

func Equals(a Value, b Value) bool {
	switch a_ := a.(type) {
	case Map:
		if bMap, ok := b.(Map); ok {
			// Does A have all the keys that are in B?
			for key := range bMap {
				if _, ok := a_[key]; !ok {
					return false
				}
			}

			// Are all values in A equal to those in B?
			for key, aValue := range a_ {
				if bValue, ok := bMap[key]; ok {
					if !Equals(aValue, bValue) {
						return false
					}
				} else {
					return false
				}
			}

			return true
		} else {
			return false
		}

	case StringMap:
		if bMap, ok := b.(StringMap); ok {
			// Does A have all the keys that are in B?
			for key := range bMap {
				if _, ok := a_[key]; !ok {
					return false
				}
			}

			// Are all values in A equal to those in B?
			for key, aValue := range a_ {
				if bValue, ok := bMap[key]; ok {
					if !Equals(aValue, bValue) {
						return false
					}
				} else {
					return false
				}
			}

			return true
		} else {
			return false
		}

	case List:
		if bList, ok := b.(List); ok {
			// Must have same lengths
			if len(a_) != len(bList) {
				return false
			}

			for index, aValue := range a_ {
				bValue := bList[index]
				if !Equals(aValue, bValue) {
					return false
				}
			}

			return true
		} else {
			return false
		}

	default:
		return a == b
	}
}
