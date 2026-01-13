package wasmtime

import (
	"math"
)

// Value encoding and decoding functions for WebAssembly values.
// All values are represented as uint64 in the API, matching wazero's design.

// EncodeI32 encodes an int32 value as uint64.
// This matches wazero's api.EncodeI32 function.
func EncodeI32(v int32) uint64 {
	return uint64(uint32(v))
}

// DecodeI32 decodes a uint64 value to int32.
func DecodeI32(v uint64) int32 {
	return int32(uint32(v))
}

// EncodeU32 encodes a uint32 value as uint64.
// This matches wazero's api.EncodeU32 function.
func EncodeU32(v uint32) uint64 {
	return uint64(v)
}

// DecodeU32 decodes a uint64 value to uint32.
func DecodeU32(v uint64) uint32 {
	return uint32(v)
}

// EncodeI64 encodes an int64 value as uint64.
// This matches wazero's api.EncodeI64 function.
func EncodeI64(v int64) uint64 {
	return uint64(v)
}

// DecodeI64 decodes a uint64 value to int64.
func DecodeI64(v uint64) int64 {
	return int64(v)
}

// EncodeF32 encodes a float32 value as uint64.
// This matches wazero's api.EncodeF32 function.
func EncodeF32(v float32) uint64 {
	return uint64(math.Float32bits(v))
}

// DecodeF32 decodes a uint64 value to float32.
func DecodeF32(v uint64) float32 {
	return math.Float32frombits(uint32(v))
}

// EncodeF64 encodes a float64 value as uint64.
// This matches wazero's api.EncodeF64 function.
func EncodeF64(v float64) uint64 {
	return math.Float64bits(v)
}

// DecodeF64 decodes a uint64 value to float64.
func DecodeF64(v uint64) float64 {
	return math.Float64frombits(v)
}

// EncodeExternref encodes a uintptr (pointer) value as an externref for passing to WebAssembly.
// This matches wazero's api.EncodeExternref function.
// Externrefs are opaque host references that can be passed between the host and WebAssembly.
func EncodeExternref(v uintptr) uint64 {
	return uint64(v)
}

// DecodeExternref decodes a uint64 value from WebAssembly to uintptr (externref).
// This matches wazero's api.DecodeExternref function.
// The returned value is an opaque host reference.
func DecodeExternref(v uint64) uintptr {
	return uintptr(v)
}
