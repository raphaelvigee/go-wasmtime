package wasmtime

import (
	"context"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/rvigee/purego-wasmtime/api"
)

// module implements api.Module for an instantiated WebAssembly module.
type module struct {
	inst  wasmtime_instance_t
	store wasmtime_store_t
	name  string
}

func (m *module) Name() string {
	return m.name
}

func (m *module) ExportedFunction(name string) api.Function {
	ext, err := m.getExport(name)
	if err != nil {
		return nil
	}

	if ext.kind != WASMTIME_EXTERN_FUNC {
		return nil
	}

	funcPtr := ext.AsFunc()
	return &function{
		name:    name,
		ptr:     funcPtr,
		store:   m.store,
		fnCache: nil, // Will be populated on first Definition() call
	}
}

func (m *module) ExportedFunctionDefinitions() map[string]api.FunctionDefinition {
	// TODO: Implement module export iteration
	// For now, return empty map
	return make(map[string]api.FunctionDefinition)
}

func (m *module) Close(ctx context.Context) error {
	// Instance doesn't need explicit cleanup in wasmtime
	// Resources are managed by the store
	return nil
}

func (m *module) getExport(name string) (*wasmtime_extern_t, error) {
	ext := new(wasmtime_extern_t)
	nameByte := []byte(name + "\000")

	storeCtx := wasmtime_store_context(m.store)
	ok := wasmtime_instance_export_get(storeCtx, &m.inst, &nameByte[0], uintptr(len(name)), ext)

	runtime.KeepAlive(m)
	runtime.KeepAlive(name)

	if !ok {
		return nil, fmt.Errorf("export %q not found", name)
	}

	return ext, nil
}

// function implements api.Function.
type function struct {
	name    string
	ptr     *wasmtime_func_t
	store   wasmtime_store_t
	fnCache api.FunctionDefinition
}

func (f *function) Definition() api.FunctionDefinition {
	if f.fnCache != nil {
		return f.fnCache
	}

	storeCtx := wasmtime_store_context(f.store)
	funcType := wasmtime_func_type(storeCtx, f.ptr)
	if funcType == 0 {
		return &functionDefinition{
			name:        f.name,
			paramTypes:  []api.ValueType{},
			resultTypes: []api.ValueType{},
		}
	}
	defer wasm_functype_delete(funcType)

	paramsVec := wasm_functype_params(funcType)
	resultsVec := wasm_functype_results(funcType)

	paramTypes := make([]api.ValueType, paramsVec.size)
	for i := uintptr(0); i < paramsVec.size; i++ {
		vt := (*wasm_valtype_t)(unsafe.Add(unsafe.Pointer(paramsVec.data), i*unsafe.Sizeof(uintptr(0))))
		paramTypes[i] = wasmValueTypeToAPI(wasm_valtype_kind(*vt))
	}

	resultTypes := make([]api.ValueType, resultsVec.size)
	for i := uintptr(0); i < resultsVec.size; i++ {
		vt := (*wasm_valtype_t)(unsafe.Add(unsafe.Pointer(resultsVec.data), i*unsafe.Sizeof(uintptr(0))))
		resultTypes[i] = wasmValueTypeToAPI(wasm_valtype_kind(*vt))
	}

	f.fnCache = &functionDefinition{
		name:        f.name,
		paramTypes:  paramTypes,
		resultTypes: resultTypes,
	}

	return f.fnCache
}

func (f *function) Call(ctx context.Context, params ...uint64) ([]uint64, error) {
	def := f.Definition()
	paramTypes := def.ParamTypes()
	resultTypes := def.ResultTypes()

	if len(params) != len(paramTypes) {
		return nil, fmt.Errorf("expected %d parameters, got %d", len(paramTypes), len(params))
	}

	// Convert uint64 params to wasmtime_val_t
	wasmParams := make([]wasmtime_val_t, len(params))
	for i, param := range params {
		wasmParams[i] = encodeToWasmValue(param, paramTypes[i])
	}

	// Prepare result buffer
	var resultsPtr *wasmtime_val_t
	var results []wasmtime_val_t
	numResults := uintptr(len(resultTypes))
	if numResults > 0 {
		results = make([]wasmtime_val_t, numResults)
		resultsPtr = &results[0]
	}

	// Call the function
	var paramsPtr *wasmtime_val_t
	if len(wasmParams) > 0 {
		paramsPtr = &wasmParams[0]
	}

	var trap *wasm_trap_t
	storeCtx := wasmtime_store_context(f.store)
	callErr := wasmtime_func_call(storeCtx, f.ptr, paramsPtr, uintptr(len(wasmParams)), resultsPtr, numResults, &trap)

	runtime.KeepAlive(f)
	runtime.KeepAlive(wasmParams)
	runtime.KeepAlive(results)

	if callErr != 0 {
		err := getErrorMessage(callErr, 0)
		// Handle WASI exit(0) gracefully
		if exitErr, ok := err.(*WASIExitError); ok && exitErr.ExitCode == 0 {
			// Success exit - return results normally
		} else {
			return nil, fmt.Errorf("call failed: %w", err)
		}
	}
	if trap != nil {
		return nil, fmt.Errorf("call failed (trap): %w", getErrorMessage(0, *trap))
	}

	// Convert results back to uint64
	resultValues := make([]uint64, numResults)
	for i := uintptr(0); i < numResults; i++ {
		resultValues[i] = decodeFromWasmValue(&results[i])
	}

	return resultValues, nil
}

// functionDefinition implements api.FunctionDefinition.
type functionDefinition struct {
	name        string
	paramTypes  []api.ValueType
	resultTypes []api.ValueType
}

func (fd *functionDefinition) Name() string {
	return fd.name
}

func (fd *functionDefinition) ParamTypes() []api.ValueType {
	return fd.paramTypes
}

func (fd *functionDefinition) ResultTypes() []api.ValueType {
	return fd.resultTypes
}

func (fd *functionDefinition) ParamNames() []string {
	// Not available in wasmtime C API
	return nil
}

func (fd *functionDefinition) ResultNames() []string {
	// Not available in wasmtime C API
	return nil
}

// Helper functions for value conversion

func wasmValueTypeToAPI(kind wasm_valkind_t) api.ValueType {
	switch kind {
	case WASM_I32:
		return api.ValueTypeI32
	case WASM_I64:
		return api.ValueTypeI64
	case WASM_F32:
		return api.ValueTypeF32
	case WASM_F64:
		return api.ValueTypeF64
	case WASM_V128:
		return api.ValueTypeV128
	case WASM_FUNCREF:
		return api.ValueTypeFuncref
	case WASM_EXTERNREF:
		return api.ValueTypeExternref
	default:
		return api.ValueTypeI32 // Default fallback
	}
}

func encodeToWasmValue(v uint64, vt api.ValueType) wasmtime_val_t {
	var result wasmtime_val_t
	// Zero-init padding
	*(*[24]byte)(unsafe.Pointer(&result)) = [24]byte{}

	switch vt {
	case api.ValueTypeI32:
		result.SetI32(DecodeI32(v))
	case api.ValueTypeI64:
		result.SetI64(DecodeI64(v))
	case api.ValueTypeF32:
		result.SetF32(DecodeF32(v))
	case api.ValueTypeF64:
		result.SetF64(DecodeF64(v))
	case api.ValueTypeV128:
		// V128 requires special handling - for now just set as i64
		result.SetI64(int64(v))
	case api.ValueTypeFuncref:
		// Funcref - for now, just pass through as zero
		// TODO: Proper funcref handling would require more complex support
		result.SetFuncRef(wasmtime_func_t{})
	case api.ValueTypeExternref:
		// Externref - pass through as direct uintptr
		result.SetExternRef(uintptr(v))
	}

	return result
}

func decodeFromWasmValue(v *wasmtime_val_t) uint64 {
	switch v.kind {
	case WASM_I32:
		return EncodeI32(v.GetI32())
	case WASM_I64:
		return EncodeI64(v.GetI64())
	case WASM_F32:
		return EncodeF32(v.GetF32())
	case WASM_F64:
		return EncodeF64(v.GetF64())
	case WASM_V128:
		// V128 requires special handling
		return uint64(v.GetI64())
	case WASM_FUNCREF:
		// Funcref - return zero for now
		// TODO: Proper funcref handling
		return 0
	case WASM_EXTERNREF:
		return uint64(v.GetExternRef())
	default:
		return 0
	}
}
