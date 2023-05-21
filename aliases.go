package ard

// Note: These are just convenient aliases, *not* types. Extra types would ensure more strictness
// but would make life more complicated than it needs to be. That said, if we *do* want to make
// these into types, then we need to make sure not to add any methods to them, otherwise the
// goja JavaScript engine will treat them as host objects instead of regular JavaScript dict
// objects.

type Value = any

type List = []Value

type Map = map[Value]Value

type StringMap = map[string]Value
