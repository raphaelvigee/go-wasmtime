package wasmtime

import (
	"fmt"
	"runtime"
	"unsafe"
)

// Instance represents an instantiated WebAssembly module
type Instance struct {
	inst   wasmtime_instance_t
	store  *Store
	module *Module
}

// NewInstance creates a new instance from a module
func NewInstance(store *Store, module *Module, imports []wasmtime_extern_t) (*Instance, error) {
	var inst wasmtime_instance_t
	var trap *wasm_trap_t

	// Convert imports to unsafe pointer (like wasmtime-go does)
	var importsPtr *wasmtime_extern_t
	if len(imports) > 0 {
		importsPtr = (*wasmtime_extern_t)(unsafe.Pointer(&imports[0]))
	}

	err := wasmtime_instance_new(store.Context(), module.ptr, importsPtr, uintptr(len(imports)), &inst, &trap)

	// Keep objects alive during C call (critical!)
	runtime.KeepAlive(store)
	runtime.KeepAlive(module)
	runtime.KeepAlive(imports)

	if err != 0 {
		return nil, fmt.Errorf("failed to instantiate: %s", getErrorMessage(err, 0))
	}
	if trap != nil {
		return nil, fmt.Errorf("failed to instantiate (trap): %s", getErrorMessage(0, *trap))
	}

	instance := &Instance{
		inst:   inst,
		store:  store,
		module: module,
	}
	return instance, nil
}

// GetExport gets an exported item by name
func (i *Instance) GetExport(name string) (*wasmtime_extern_t, error) {
	var ext wasmtime_extern_t
	nameBytes := []byte(name)

	ok := wasmtime_instance_export_get(i.store.Context(), &i.inst, &nameBytes[0], uintptr(len(name)), &ext)

	runtime.KeepAlive(i.store)
	runtime.KeepAlive(name)

	if !ok {
		return nil, fmt.Errorf("export %q not found", name)
	}

	return &ext, nil
}

// Call calls an exported function by name
func (i *Instance) Call(name string, args ...interface{}) ([]interface{}, error) {
	// Get the exported function
	ext, err := i.GetExport(name)
	if err != nil {
		return nil, err
	}

	if ext.kind != WASMTIME_EXTERN_FUNC {
		return nil, fmt.Errorf("export %q is not a function", name)
	}

	// Convert args to wasmtime values
	var wasmArgs []wasmtime_val_t
	for _, arg := range args {
		val, err := toWasmValue(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid argument: %w", err)
		}
		wasmArgs = append(wasmArgs, val)
	}

	// Call the function (for now, assume no results for simplicity)
	var trap *wasm_trap_t
	var argsPtr *wasmtime_val_t
	if len(wasmArgs) > 0 {
		argsPtr = (*wasmtime_val_t)(unsafe.Pointer(&wasmArgs[0]))
	}

	funcPtr := ext.AsFunc()
	callErr := wasmtime_func_call(i.store.Context(), funcPtr, argsPtr, uintptr(len(wasmArgs)), nil, 0, &trap)

	// Keep objects alive during C call
	runtime.KeepAlive(i.store)
	runtime.KeepAlive(name)
	runtime.KeepAlive(wasmArgs)

	if callErr != 0 {
		return nil, fmt.Errorf("call failed: %s", getErrorMessage(callErr, 0))
	}
	if trap != nil {
		return nil, fmt.Errorf("call failed (trap): %s", getErrorMessage(0, *trap))
	}

	return nil, nil
}

// toWasmValue converts a Go value to a wasmtime value
func toWasmValue(v interface{}) (wasmtime_val_t, error) {
	switch val := v.(type) {
	case int32:
		return wasmtime_val_t{kind: 0, of: wasmtime_val_raw{i64: int64(val)}}, nil
	case int64:
		return wasmtime_val_t{kind: 1, of: wasmtime_val_raw{i64: val}}, nil
	case int:
		return wasmtime_val_t{kind: 0, of: wasmtime_val_raw{i64: int64(val)}}, nil
	default:
		return wasmtime_val_t{}, fmt.Errorf("unsupported type: %T", v)
	}
}
