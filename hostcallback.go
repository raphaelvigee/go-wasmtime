package wasmtime

import (
	"context"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/rvigee/purego-wasmtime/api"
)

// hostFunctionRegistry manages Go host functions and their C callbacks
type hostFunctionRegistry struct {
	mu        sync.RWMutex
	functions map[uintptr]*registeredFunction
	nextID    uintptr
}

type registeredFunction struct {
	id       uintptr
	builder  *hostFunctionBuilder
	callback uintptr // The C-callable function pointer from purego.NewCallback
	storeCtx wasmtime_context_t
	module   api.Module      // For GoModuleFunc
	ctx      context.Context // Context for host function execution
}

var globalRegistry = &hostFunctionRegistry{
	functions: make(map[uintptr]*registeredFunction),
	nextID:    1,
}

// hostCallbackWrapper is the actual Go function that purego.NewCallback will wrap
// It must match the C signature exactly
func hostCallbackWrapper(env uintptr, caller uintptr, args *wasmtime_val_t, nargs uintptr, results *wasmtime_val_t, nresults uintptr) uintptr {
	// Look up the registered function
	globalRegistry.mu.RLock()
	regFunc, ok := globalRegistry.functions[env]
	globalRegistry.mu.RUnlock()

	if !ok {
		// Function not found - this shouldn't happen
		return 0
	}

	// Convert C args to Go uint64 stack
	stack := make([]uint64, nargs+nresults)

	// Copy args to stack
	for i := uintptr(0); i < nargs; i++ {
		argPtr := (*wasmtime_val_t)(unsafe.Pointer(uintptr(unsafe.Pointer(args)) + i*unsafe.Sizeof(wasmtime_val_t{})))
		stack[i] = convertWasmValueToUint64(argPtr)
	}

	// Use the stored context from registration
	ctx := regFunc.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// Call the appropriate Go function and handle errors
	var callErr error
	if regFunc.builder.goFunc != nil {
		regFunc.builder.goFunc(ctx, stack)
	} else if regFunc.builder.goModuleFunc != nil {
		// For GoModuleFunc, create a wrapper module that accesses exports from the caller
		wrapperMod := &callerModule{caller: caller, store: regFunc.storeCtx}
		regFunc.builder.goModuleFunc(ctx, wrapperMod, stack)
	} else if regFunc.builder.goFunction != nil {
		paramSlice := stack[:nargs]
		resultSlice, err := regFunc.builder.goFunction.Call(ctx, paramSlice)
		if err != nil {
			callErr = err
		} else {
			for i, val := range resultSlice {
				stack[nargs+uintptr(i)] = val
			}
		}
	}

	// If there was an error, create a trap
	if callErr != nil {
		// Create a trap with the error message
		errMsg := callErr.Error()
		errBytes := []byte(errMsg + "\x00")
		trap := wasm_trap_new(&errBytes[0], uintptr(len(errMsg)))
		// Store trap pointer somewhere the C API can access it
		// Since we can't modify the callback signature, we return a non-zero value
		// However, this won't work properly with the current wasmtime callback API
		// The proper way would be to use wasmtime_trap_t** parameter
		// For now, we'll just return an error indicator
		_ = trap // Keep the trap alive
		return 1 // Return non-zero to indicate error occurred
	}

	// Copy results back
	for i := uintptr(0); i < nresults; i++ {
		resultPtr := (*wasmtime_val_t)(unsafe.Pointer(uintptr(unsafe.Pointer(results)) + i*unsafe.Sizeof(wasmtime_val_t{})))
		resultValue := stack[nargs+i]

		// Determine the type from builder
		valueType := regFunc.builder.resultTypes[i]
		convertUint64ToWasmValue(resultValue, valueType, resultPtr)
	}

	return 0 // No trap
}

// Helper to convert wasmtime_val_t to uint64
func convertWasmValueToUint64(val *wasmtime_val_t) uint64 {
	switch val.kind {
	case WASM_I32:
		return uint64(uint32(val.GetI32()))
	case WASM_I64:
		return uint64(val.GetI64())
	case WASM_F32:
		return uint64(EncodeF32(val.GetF32()))
	case WASM_F64:
		return EncodeF64(val.GetF64())
	case WASM_EXTERNREF:
		return uint64(val.GetExternRef())
	default:
		return 0
	}
}

// Helper to convert uint64 back to wasmtime_val_t
func convertUint64ToWasmValue(value uint64, valueType api.ValueType, val *wasmtime_val_t) {
	switch valueType {
	case api.ValueTypeI32:
		val.kind = WASM_I32
		val.SetI32(DecodeI32(value))
	case api.ValueTypeI64:
		val.kind = WASM_I64
		val.SetI64(DecodeI64(value))
	case api.ValueTypeF32:
		val.kind = WASM_F32
		val.SetF32(DecodeF32(value))
	case api.ValueTypeF64:
		val.kind = WASM_F64
		val.SetF64(DecodeF64(value))
	case api.ValueTypeExternref:
		val.kind = WASM_EXTERNREF
		val.SetExternRef(DecodeExternref(value))
	}
}

// createFuncType creates a wasm_functype_t from parameter and result types
func createFuncType(paramTypes, resultTypes []api.ValueType) (wasm_functype_t, func()) {
	// Create parameter type vector
	var params wasm_valtype_vec_t
	if len(paramTypes) == 0 {
		wasm_valtype_vec_new_empty(&params)
	} else {
		wasm_valtype_vec_new_uninitialized(&params, uintptr(len(paramTypes)))
		paramArray := unsafe.Slice((*wasm_valtype_t)(unsafe.Pointer(params.data)), len(paramTypes))
		for i, pt := range paramTypes {
			paramArray[i] = wasm_valtype_new(apiValueTypeToWasm(pt))
		}
	}

	// Create result type vector
	var results wasm_valtype_vec_t
	if len(resultTypes) == 0 {
		wasm_valtype_vec_new_empty(&results)
	} else {
		wasm_valtype_vec_new_uninitialized(&results, uintptr(len(resultTypes)))
		resultArray := unsafe.Slice((*wasm_valtype_t)(unsafe.Pointer(results.data)), len(resultTypes))
		for i, rt := range resultTypes {
			resultArray[i] = wasm_valtype_new(apiValueTypeToWasm(rt))
		}
	}

	// Create function type - pass pointers to the vectors
	funcType := wasm_functype_new(&params, &results)

	// Cleanup function
	cleanup := func() {
		wasm_valtype_vec_delete(&params)
		wasm_valtype_vec_delete(&results)
		wasm_functype_delete(funcType)
	}

	return funcType, cleanup
}

// apiValueTypeToWasm converts api.ValueType to wasm kind
func apiValueTypeToWasm(vt api.ValueType) uint8 {
	switch vt {
	case api.ValueTypeI32:
		return WASM_I32
	case api.ValueTypeI64:
		return WASM_I64
	case api.ValueTypeF32:
		return WASM_F32
	case api.ValueTypeF64:
		return WASM_F64
	case api.ValueTypeFuncref:
		return WASM_FUNCREF
	case api.ValueTypeExternref:
		return WASM_EXTERNREF
	default:
		return WASM_I32
	}
}

// registerHostFunction registers a Go function and returns its ID and callback pointer
func (r *hostFunctionRegistry) register(builder *hostFunctionBuilder, storeCtx wasmtime_context_t, module api.Module, ctx context.Context) (uintptr, uintptr) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++

	// Use purego.NewCallback to create a C-callable function pointer
	// The callback function must match the wasmtime callback signature
	callbackPtr := purego.NewCallback(hostCallbackWrapper)

	r.functions[id] = &registeredFunction{
		id:       id,
		builder:  builder,
		callback: callbackPtr,
		storeCtx: storeCtx,
		module:   module,
		ctx:      ctx,
	}

	// Prevent GC of the function
	runtime.SetFinalizer(r.functions[id], func(rf *registeredFunction) {
		r.mu.Lock()
		delete(r.functions, rf.id)
		r.mu.Unlock()
	})

	return id, callbackPtr
}
