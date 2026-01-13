package api

import (
	"context"
	"unsafe"
)

// Module is an instantiated WebAssembly module.
type Module interface {
	// Name returns the name of the module.
	Name() string

	// ExportedFunction returns an exported function by name, or nil if not found.
	ExportedFunction(name string) Function

	// ExportedFunctionDefinitions returns all exported function definitions.
	ExportedFunctionDefinitions() map[string]FunctionDefinition

	// ExportedMemory returns an exported memory by name, or nil if not found.
	ExportedMemory(name string) Memory

	// ExportedGlobal returns an exported global by name, or nil if not found.
	ExportedGlobal(name string) Global

	// ExportedTable returns an exported table by name, or nil if not found.
	ExportedTable(name string) Table

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

// Memory is an exported WebAssembly memory.
type Memory interface {
	// Data returns a pointer to the beginning of the memory.
	Data(ctx context.Context) unsafe.Pointer

	// DataSize returns the size of the memory in bytes.
	DataSize(ctx context.Context) uintptr

	// Size returns the size of the memory in pages.
	Size(ctx context.Context) uint64

	// Grow grows the memory by the given number of pages.
	// Returns the previous size in pages, or false if failed.
	Grow(ctx context.Context, delta uint64) (uint64, bool)
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

// Closer is a general interface for resources that need explicit cleanup.
// This matches wazero's api.Closer interface.
type Closer interface {
	// Close closes the resource and releases associated resources.
	Close(ctx context.Context) error
}

// Global is an exported WebAssembly global variable.
// This matches wazero's api.Global interface.
type Global interface {
	// Type returns the type of the global variable.
	Type() ValueType

	// Get returns the current value of the global as uint64.
	Get(ctx context.Context) uint64

	// Set sets the value of the global if it is mutable.
	// Returns an error if the global is immutable.
	Set(ctx context.Context, v uint64) error
}

// Table is an exported WebAssembly table.
// This matches wazero's api.Table interface.
type Table interface {
	// Type returns the element type of the table.
	Type(ctx context.Context) ValueType

	// Size returns the current size of the table.
	Size(ctx context.Context) uint32

	// Grow grows the table by delta elements.
	// Returns the previous size, or false if the operation failed.
	Grow(ctx context.Context, delta uint32) (uint32, bool)

	// Get retrieves the element at the given index.
	// Returns 0 if the index is out of bounds.
	Get(ctx context.Context, index uint32) uint64

	// Set sets the element at the given index.
	// Returns an error if the index is out of bounds.
	Set(ctx context.Context, index uint32, v uint64) error
}

// MemoryDefinition describes a memory's limits and characteristics.
// This matches wazero's api.MemoryDefinition interface.
type MemoryDefinition interface {
	// Min returns the minimum memory size in pages (64KiB each).
	Min() uint32

	// Max returns the maximum memory size in pages, or 0 if unbounded.
	Max() uint32

	// IsMaxEncoded returns true if a maximum size was encoded in the binary.
	IsMaxEncoded() bool
}

// ExportDefinition describes an exported item from a module.
// This matches wazero's api.ExportDefinition interface.
type ExportDefinition interface {
	// ModuleName returns the module name (empty for the current module).
	ModuleName() string

	// Name returns the export name.
	Name() string

	// Type returns the type of the export (function, table, memory, or global).
	Type() ExternType
}

// ExternType represents the type of an external item.
type ExternType byte

const (
	// ExternTypeFunc represents a function export/import.
	ExternTypeFunc ExternType = iota
	// ExternTypeGlobal represents a global export/import.
	ExternTypeGlobal
	// ExternTypeTable represents a table export/import.
	ExternTypeTable
	// ExternTypeMemory represents a memory export/import.
	ExternTypeMemory
)

// String returns the string representation of the extern type.
func (e ExternType) String() string {
	switch e {
	case ExternTypeFunc:
		return "func"
	case ExternTypeGlobal:
		return "global"
	case ExternTypeTable:
		return "table"
	case ExternTypeMemory:
		return "memory"
	default:
		return "unknown"
	}
}

// CustomSection represents a custom section in a WebAssembly module.
// This matches wazero's api.CustomSection type.
type CustomSection struct {
	// Name is the name of the custom section.
	Name string

	// Data is the raw bytes of the custom section.
	Data []byte
}
