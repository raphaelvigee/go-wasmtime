package wasmtime

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/rvigee/purego-wasmtime/api"
)

type callBuffer struct {
	Params  []wasmtime_val_t
	Results []wasmtime_val_t
	Trap    *wasm_trap_t
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &callBuffer{
			Params:  make([]wasmtime_val_t, 0, 8),
			Results: make([]wasmtime_val_t, 0, 1),
		}
	},
}

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
	f := &function{
		name:     name,
		val:      *funcPtr,
		store:    m.store,
		storeCtx: 0, // will be set below
		fnCache:  nil,
	}

	// Pre-populate definition to cache types
	def := f.Definition()
	f.paramTypes = def.ParamTypes()
	f.resultTypes = def.ResultTypes()
	f.storeCtx = wasmtime_store_context(f.store)

	return f
}

func (m *module) ExportedFunctionDefinitions() map[string]api.FunctionDefinition {
	// Wasmtime C API limitation: Export iteration APIs are not available.
	// Workaround: Call ExportedFunction(name) directly for known function names.
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

func (m *module) ExportedMemory(name string) api.Memory {
	ext, err := m.getExport(name)
	if err != nil {
		return nil
	}

	if ext.kind != WASMTIME_EXTERN_MEMORY {
		return nil
	}

	memPtr := ext.AsMemory()
	return &memory{
		val:      *memPtr,
		store:    m.store,
		storeCtx: wasmtime_store_context(m.store),
	}
}

func (m *module) ExportedGlobal(name string) api.Global {
	ext, err := m.getExport(name)
	if err != nil || ext == nil {
		return nil
	}

	if ext.kind != WASMTIME_EXTERN_GLOBAL {
		return nil
	}

	globalPtr := ext.AsGlobal()
	return &global{
		val:      *globalPtr,
		store:    m.store,
		storeCtx: wasmtime_store_context(m.store),
		// Wasmtime C API limitation: No wasmtime_global_type() or similar API available.
		// Type and mutability cannot be queried, so we use safe defaults.
		valType: api.ValueTypeI32, // Default - actual type depends on the global
		mutable: true,             // Default - assume mutable for safety
	}
}

func (m *module) ExportedTable(name string) api.Table {
	ext, err := m.getExport(name)
	if err != nil || ext == nil {
		return nil
	}

	if ext.kind != WASMTIME_EXTERN_TABLE {
		return nil
	}

	tablePtr := ext.AsTable()
	return &table{
		val:      *tablePtr,
		store:    m.store,
		storeCtx: wasmtime_store_context(m.store),
	}
}

type memory struct {
	val      wasmtime_memory_t
	store    wasmtime_store_t
	storeCtx wasmtime_context_t
}

func (m *memory) Data(ctx context.Context) unsafe.Pointer {
	return wasmtime_memory_data(m.storeCtx, &m.val)
}

func (m *memory) DataSize(ctx context.Context) uintptr {
	return wasmtime_memory_data_size(m.storeCtx, &m.val)
}

func (m *memory) Size(ctx context.Context) uint64 {
	return wasmtime_memory_size(m.storeCtx, &m.val)
}

func (m *memory) Grow(ctx context.Context, delta uint64) (uint64, bool) {
	var prevSize uint64
	err := wasmtime_memory_grow(m.storeCtx, &m.val, delta, &prevSize)
	if err != 0 {
		wasmtime_error_delete(err)
		return 0, false
	}
	return prevSize, true
}

// function implements api.Function.
type function struct {
	name        string
	val         wasmtime_func_t
	store       wasmtime_store_t
	storeCtx    wasmtime_context_t
	fnCache     api.FunctionDefinition
	paramTypes  []api.ValueType
	resultTypes []api.ValueType
}

func (f *function) Definition() api.FunctionDefinition {
	if f.fnCache != nil {
		return f.fnCache
	}

	storeCtx := wasmtime_store_context(f.store)
	funcType := wasmtime_func_type(storeCtx, &f.val)
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
	if len(params) != len(f.paramTypes) {
		return nil, fmt.Errorf("expected %d parameters, got %d", len(f.paramTypes), len(params))
	}

	// Get buffer from pool
	buf := bufferPool.Get().(*callBuffer)
	defer bufferPool.Put(buf)

	// Prepare params
	buf.Params = buf.Params[:len(params)]
	for i, param := range params {
		// Inline encodeToWasmValue
		// Optimization: Skip zero-init of padding since we override the value anyway
		// and C API ignores padding.
		vt := f.paramTypes[i]
		switch vt {
		case api.ValueTypeI32:
			buf.Params[i].kind = WASM_I32
			*(*int32)(unsafe.Pointer(&buf.Params[i].of.data[0])) = int32(param)
		case api.ValueTypeI64:
			buf.Params[i].kind = WASM_I64
			*(*int64)(unsafe.Pointer(&buf.Params[i].of.data[0])) = int64(param)
		case api.ValueTypeF32:
			buf.Params[i].kind = WASM_F32
			*(*float32)(unsafe.Pointer(&buf.Params[i].of.data[0])) = *(*float32)(unsafe.Pointer(&param))
		case api.ValueTypeF64:
			buf.Params[i].kind = WASM_F64
			*(*float64)(unsafe.Pointer(&buf.Params[i].of.data[0])) = *(*float64)(unsafe.Pointer(&param))
		case api.ValueTypeV128:
			buf.Params[i].kind = WASM_V128
			*(*int64)(unsafe.Pointer(&buf.Params[i].of.data[0])) = int64(param)
		case api.ValueTypeFuncref:
			buf.Params[i].kind = WASM_FUNCREF
			*(*wasmtime_func_t)(unsafe.Pointer(&buf.Params[i].of.data[0])) = wasmtime_func_t{}
		case api.ValueTypeExternref:
			buf.Params[i].kind = WASM_EXTERNREF
			*(*uintptr)(unsafe.Pointer(&buf.Params[i].of.data[0])) = uintptr(param)
		}
	}

	// Prepare results
	numResults := len(f.resultTypes)
	buf.Results = buf.Results[:0]
	if cap(buf.Results) < numResults {
		buf.Results = make([]wasmtime_val_t, 0, numResults)
	}
	// We need the slice length to be numResults for the C API to write into
	buf.Results = buf.Results[:numResults]

	var resultsPtr *wasmtime_val_t
	if numResults > 0 {
		resultsPtr = &buf.Results[0]
	}

	var paramsPtr *wasmtime_val_t
	if len(buf.Params) > 0 {
		paramsPtr = &buf.Params[0]
	}

	// Reset the trap pointer in the reused buffer
	buf.Trap = nil

	callErr := wasmtime_func_call(f.storeCtx, &f.val, paramsPtr, uintptr(len(buf.Params)), resultsPtr, uintptr(numResults), &buf.Trap)

	runtime.KeepAlive(f)
	// buf is kept alive by the function scope reference

	if callErr != 0 {
		err := getErrorMessage(callErr, 0)
		// Handle WASI exit(0) gracefully
		if exitErr, ok := err.(*WASIExitError); ok && exitErr.ExitCode == 0 {
			// Success exit - return results normally
		} else {
			return nil, fmt.Errorf("call failed: %w", err)
		}
	}
	if buf.Trap != nil {
		return nil, fmt.Errorf("call failed (trap): %w", getErrorMessage(0, *buf.Trap))
	}

	// Convert results back to uint64
	// We return a new slice because the caller owns the result
	// Assuming the caller wants []uint64 and not a reused buffer
	// Optimizing this further would require changing the API to accept a result buffer
	resultValues := make([]uint64, numResults)
	for i := 0; i < numResults; i++ {
		// Inline decodeFromWasmValue
		val := &buf.Results[i]
		switch val.kind {
		case WASM_I32:
			resultValues[i] = uint64(uint32(*(*int32)(unsafe.Pointer(&val.of.data[0]))))
		case WASM_I64:
			resultValues[i] = uint64(*(*int64)(unsafe.Pointer(&val.of.data[0])))
		case WASM_F32:
			f32 := *(*float32)(unsafe.Pointer(&val.of.data[0]))
			resultValues[i] = uint64(*(*uint32)(unsafe.Pointer(&f32)))
		case WASM_F64:
			f64 := *(*float64)(unsafe.Pointer(&val.of.data[0]))
			resultValues[i] = *(*uint64)(unsafe.Pointer(&f64))
		case WASM_V128:
			resultValues[i] = uint64(*(*int64)(unsafe.Pointer(&val.of.data[0])))
		case WASM_FUNCREF:
			resultValues[i] = 0
		case WASM_EXTERNREF:
			resultValues[i] = uint64(*(*uintptr)(unsafe.Pointer(&val.of.data[0])))
		default:
			resultValues[i] = 0
		}
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
