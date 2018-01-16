// Package contextkeys stores the keys to the context accessor
// functions, letting generated code safely set values in contexts
// without exposing the setters to the outside world.
package contextkeys

type contextKey int

const (
	MethodNameKey contextKey = 1 + iota
	ServiceNameKey
	PackageNameKey
	StatusCodeKey
	RequestHeaderKey
	ResponseWriterKey
)
