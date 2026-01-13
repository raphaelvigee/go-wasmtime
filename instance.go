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
		return nil, fmt.Errorf("failed to instantiate: %w", getErrorMessage(err, 0))
	}
	if trap != nil {
		return nil, fmt.Errorf("failed to instantiate (trap): %w", getErrorMessage(0, *trap))
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
	// CRITICAL: Must allocate on heap because we return a pointer!
	// Stack variables become invalid after function returns
	ext := new(wasmtime_extern_t)

	nameByte := []byte(name + "\000")
	ok := wasmtime_instance_export_get(i.store.Context(), &i.inst, &nameByte[0], uintptr(len(name)), ext)

	runtime.KeepAlive(i.store)
	runtime.KeepAlive(name)

	if !ok {
		return nil, fmt.Errorf("export %q not found", name)
	}

	return ext, nil
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

	// Call the function
	var trap *wasm_trap_t
	var argsPtr *wasmtime_val_t
	if len(wasmArgs) > 0 {
		argsPtr = (*wasmtime_val_t)(unsafe.Pointer(&wasmArgs[0]))
	}

	funcPtr := ext.AsFunc()

	// Query function type to get expected result count
	funcType := wasmtime_func_type(i.store.Context(), funcPtr)
	defer wasm_functype_delete(funcType) // Clean up the functype

	resultsVec := wasm_functype_results(funcType)
	numResults := resultsVec.size

	// Allocate result buffer dynamically based on expected results
	var resultsPtr *wasmtime_val_t
	var results []wasmtime_val_t
	if numResults > 0 {
		results = make([]wasmtime_val_t, numResults)
		resultsPtr = &results[0]
	}

	// Call the function
	callErr := wasmtime_func_call(i.store.Context(), funcPtr, argsPtr, uintptr(len(wasmArgs)), resultsPtr, numResults, &trap)

	// Keep objects alive during C call
	runtime.KeepAlive(i.store)
	runtime.KeepAlive(name)
	runtime.KeepAlive(wasmArgs)
	runtime.KeepAlive(results)

	if callErr != 0 {
		err := getErrorMessage(callErr, 0)
		// Automatically ignore WASI exit code 0 for better UX
		if exitErr, ok := err.(*WASIExitError); ok && exitErr.ExitCode == 0 {
			// WASI exit(0) is success - return results normally
		} else {
			return nil, fmt.Errorf("call failed: %w", err)
		}
	}
	if trap != nil {
		return nil, fmt.Errorf("call failed (trap): %w", getErrorMessage(0, *trap))
	}

	// Convert results back to Go values
	goResults := make([]interface{}, 0, numResults)
	for idx := uintptr(0); idx < numResults; idx++ {
		val, err := fromWasmValue(&results[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to convert result: %w", err)
		}
		goResults = append(goResults, val)
	}

	return goResults, nil
}

// toWasmValue converts a Go value to a wasmtime value
func toWasmValue(v interface{}) (wasmtime_val_t, error) {
	var result wasmtime_val_t
	// Zero-init padding to avoid garbage
	*(*[24]byte)(unsafe.Pointer(&result)) = [24]byte{}

	switch val := v.(type) {
	case int32:
		result.SetI32(val)
		return result, nil
	case int64:
		result.SetI64(val)
		return result, nil
	case int:
		result.SetI32(int32(val))
		return result, nil
	default:
		return wasmtime_val_t{}, fmt.Errorf("unsupported type: %T", v)
	}
}

// fromWasmValue converts a wasmtime value back to a Go value
func fromWasmValue(v *wasmtime_val_t) (interface{}, error) {
	switch v.kind {
	case 0: // i32
		return v.GetI32(), nil
	case 1: // i64
		return v.GetI64(), nil
	default:
		return nil, fmt.Errorf("unsupported wasm type: %d", v.kind)
	}
}
