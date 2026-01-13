package wasmtime

import (
	"context"
	"fmt"

	"github.com/rvigee/purego-wasmtime/api"
)

// GoFunc is a host function callable from WebAssembly.
// It receives a context and a stack of parameters/results encoded as uint64.
// The function should read parameters from the beginning of the stack
// and write results starting at index len(params).
//
// This matches wazero's api.GoFunc type.
type GoFunc func(ctx context.Context, stack []uint64)

// GoModuleFunc is a host function that also receives the calling module.
// This is useful for functions that need to access the module's memory or other exports.
//
// This matches wazero's api.GoModuleFunc type.
type GoModuleFunc func(ctx context.Context, mod api.Module, stack []uint64)

// GoFunction is a more ergonomic interface for defining host functions
// where parameters and results are passed as separate slices.
//
// This matches wazero's api.GoFunction type.
type GoFunction interface {
	// Call invokes the function with the given parameters and returns results.
	Call(ctx context.Context, params []uint64) ([]uint64, error)
}

// HostFunctionBuilder provides a fluent API for defining host functions.
type HostFunctionBuilder interface {
	// WithGoFunc sets the function implementation as a GoFunc.
	WithGoFunc(fn GoFunc) HostFunctionBuilder

	// WithGoModuleFunc sets the function implementation as a GoModuleFunc.
	WithGoModuleFunc(fn GoModuleFunc) HostFunctionBuilder

	// WithGoFunction sets the function implementation as a GoFunction.
	WithGoFunction(fn GoFunction) HostFunctionBuilder

	// WithParameterNames sets the parameter names (optional, for debugging).
	WithParameterNames(names ...string) HostFunctionBuilder

	// WithResultNames sets the result names (optional, for debugging).
	WithResultNames(names ...string) HostFunctionBuilder

	// Export finalizes the function and exports it with the given name.
	Export(name string)
}

// HostModuleBuilder provides a fluent API for creating host modules.
// Host modules contain functions, globals, memories defined in Go.
type HostModuleBuilder interface {
	// NewFunctionBuilder creates a new function with the given name and signature.
	// paramTypes and resultTypes should use api.ValueType constants.
	NewFunctionBuilder(name string, paramTypes, resultTypes []api.ValueType) HostFunctionBuilder

	// Instantiate creates the host module and makes it available for imports.
	// The module name should match the import name in the WASM module.
	Instantiate(ctx context.Context) error

	// Close cleans up the host module resources.
	Close(ctx context.Context) error
}

// hostFunctionBuilder implements HostFunctionBuilder
type hostFunctionBuilder struct {
	parent       *hostModuleBuilder
	name         string
	paramTypes   []api.ValueType
	resultTypes  []api.ValueType
	goFunc       GoFunc
	goModuleFunc GoModuleFunc
	goFunction   GoFunction
	paramNames   []string
	resultNames  []string
}

func (hfb *hostFunctionBuilder) WithGoFunc(fn GoFunc) HostFunctionBuilder {
	hfb.goFunc = fn
	hfb.goModuleFunc = nil
	hfb.goFunction = nil
	return hfb
}

func (hfb *hostFunctionBuilder) WithGoModuleFunc(fn GoModuleFunc) HostFunctionBuilder {
	hfb.goModuleFunc = fn
	hfb.goFunc = nil
	hfb.goFunction = nil
	return hfb
}

func (hfb *hostFunctionBuilder) WithGoFunction(fn GoFunction) HostFunctionBuilder {
	hfb.goFunction = fn
	hfb.goFunc = nil
	hfb.goModuleFunc = nil
	return hfb
}

func (hfb *hostFunctionBuilder) WithParameterNames(names ...string) HostFunctionBuilder {
	hfb.paramNames = names
	return hfb
}

func (hfb *hostFunctionBuilder) WithResultNames(names ...string) HostFunctionBuilder {
	hfb.resultNames = names
	return hfb
}

func (hfb *hostFunctionBuilder) Export(name string) {
	hfb.name = name
	hfb.parent.addFunction(hfb)
}

// hostModuleBuilder implements HostModuleBuilder
type hostModuleBuilder struct {
	moduleName string
	functions  []*hostFunctionBuilder
	runtime    *wasmRuntime
	linker     wasmtime_linker_t
}

func (hmb *hostModuleBuilder) NewFunctionBuilder(name string, paramTypes, resultTypes []api.ValueType) HostFunctionBuilder {
	return &hostFunctionBuilder{
		parent:      hmb,
		name:        name,
		paramTypes:  paramTypes,
		resultTypes: resultTypes,
	}
}

func (hmb *hostModuleBuilder) addFunction(fn *hostFunctionBuilder) {
	hmb.functions = append(hmb.functions, fn)
}

func (hmb *hostModuleBuilder) Instantiate(ctx context.Context) error {
	if hmb.runtime == nil {
		return fmt.Errorf("host module builder has no associated runtime")
	}

	storeCtx := wasmtime_store_context(hmb.runtime.store)

	// Register each function with the linker
	for _, fn := range hmb.functions {
		// Create function type
		funcType, cleanup := createFuncType(fn.paramTypes, fn.resultTypes)
		defer cleanup()

		// Register function in global registry
		funcID, callbackPtr := globalRegistry.register(fn, storeCtx, nil, ctx)

		// Get callback pointer

		// Create wasmtime function
		var wasmFunc wasmtime_func_t
		wasmtime_func_new(
			storeCtx,
			funcType,
			callbackPtr,
			funcID, // env pointer (our function ID)
			0,      // finalizer
			&wasmFunc,
		)

		// Create extern from function
		var ext wasmtime_extern_t
		ext.kind = WASMTIME_EXTERN_FUNC
		funcPtr := ext.AsFunc()
		*funcPtr = wasmFunc

		// Define in linker
		moduleBytes := []byte(hmb.moduleName + "\000")
		nameBytes := []byte(fn.name + "\000")

		err := wasmtime_linker_define(
			hmb.linker,
			storeCtx, // Add the missing store context
			&moduleBytes[0],
			uintptr(len(hmb.moduleName)),
			&nameBytes[0],
			uintptr(len(fn.name)),
			&ext,
		)

		if err != 0 {
			wasmtime_error_delete(err)
			return fmt.Errorf("failed to define host function %s::%s", hmb.moduleName, fn.name)
		}
	}

	return nil
}

func (hmb *hostModuleBuilder) Close(ctx context.Context) error {
	// Clean up resources
	return nil
}

// NewHostModuleBuilder creates a new host module with the given name.
// This function is called on the Runtime to create host modules.
func (r *wasmRuntime) NewHostModuleBuilder(name string) HostModuleBuilder {
	return &hostModuleBuilder{
		moduleName: name,
		runtime:    r,
		linker:     r.linker,
		functions:  make([]*hostFunctionBuilder, 0),
	}
}
