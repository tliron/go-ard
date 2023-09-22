package ard

// Note: These are just convenient aliases, *not* types. Extra types would ensure more
// strictness but would make life more complicated than it needs to be. That said, if we *do*
// want to make these into types, then we need to make sure not to add any methods to them,
// otherwise the goja JavaScript engine will treat them as host objects instead of regular
// JavaScript dict objects.

// An alias used to signify that ARD values are expected,
// namely primitives, [List], [Map], and [StringMap] nested to
// any depth.
type Value = any

// An alias used to signify a list of ARD [Value].
type List = []Value

// An alias used to signify a map of ARD [Value] in which the
// key is also a [Value]. Note that this type does not protect
// its users against using invalid keys. Keys must be both hashable
// comparable.
type Map = map[Value]Value

// An alias used to signify a map of ARD [Value] in which the
// key is always string. This alias is introduced for compability
// with certain parsers and encoders.
type StringMap = map[string]Value
