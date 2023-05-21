package ard

func Canonicalize(value Value) (Value, error) {
	return NewReflector().Unpack(value)
}
