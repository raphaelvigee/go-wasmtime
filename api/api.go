package api

import "context"

// Module is an instantiated WebAssembly module.
type Module interface {
	// Name returns the name of the module.
	Name() string

	// ExportedFunction returns an exported function by name, or nil if not found.
	ExportedFunction(name string) Function

	// ExportedFunctionDefinitions returns all exported function definitions.
	ExportedFunctionDefinitions() map[string]FunctionDefinition

	// Close closes the module and releases associated resources.
	Close(ctx context.Context) error
}

// Function is an exported WebAssembly function.
type Function interface {
	// Definition returns the function signature.
	Definition() FunctionDefinition

	// Call invokes the function with the given parameters.
	// Parameters and results are encoded as uint64 values.
	Call(ctx context.Context, params ...uint64) ([]uint64, error)
}

// FunctionDefinition describes a function's signature.
type FunctionDefinition interface {
	// Name returns the function name.
	Name() string

	// ParamTypes returns the types of the function parameters.
	ParamTypes() []ValueType

	// ResultTypes returns the types of the function results.
	ResultTypes() []ValueType

	// ParamNames returns the names of the function parameters, if available.
	ParamNames() []string

	// ResultNames returns the names of the function results, if available.
	ResultNames() []string
}

// ValueType is a WebAssembly value type.
type ValueType byte

const (
	// ValueTypeI32 is a 32-bit integer.
	ValueTypeI32 ValueType = iota
	// ValueTypeI64 is a 64-bit integer.
	ValueTypeI64
	// ValueTypeF32 is a 32-bit float.
	ValueTypeF32
	// ValueTypeF64 is a 64-bit float.
	ValueTypeF64
	// ValueTypeV128 is a 128-bit vector.
	ValueTypeV128
	// ValueTypeFuncref is a function reference.
	ValueTypeFuncref
	// ValueTypeExternref is an external reference.
	ValueTypeExternref
)

// String returns the string representation of the value type.
func (v ValueType) String() string {
	switch v {
	case ValueTypeI32:
		return "i32"
	case ValueTypeI64:
		return "i64"
	case ValueTypeF32:
		return "f32"
	case ValueTypeF64:
		return "f64"
	case ValueTypeV128:
		return "v128"
	case ValueTypeFuncref:
		return "funcref"
	case ValueTypeExternref:
		return "externref"
	default:
		return "unknown"
	}
}
