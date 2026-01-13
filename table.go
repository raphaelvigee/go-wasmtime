package wasmtime

import (
	"context"
	"fmt"

	"github.com/rvigee/purego-wasmtime/api"
)

// table implements api.Table for a WebAssembly table.
type table struct {
	val      wasmtime_table_t
	store    wasmtime_store_t
	storeCtx wasmtime_context_t
}

func (t *table) Type(ctx context.Context) api.ValueType {
	// Wasmtime C API limitation: No wasmtime_table_type() API available.
	// Defaulting to funcref as it's the most common table element type.
	return api.ValueTypeFuncref
}

func (t *table) Size(ctx context.Context) uint32 {
	return wasmtime_table_size(t.storeCtx, &t.val)
}

func (t *table) Grow(ctx context.Context, delta uint32) (uint32, bool) {
	var prevSize uint32
	// Initialize value for growth (typically null for funcref)
	var val wasmtime_val_t
	val.kind = WASM_FUNCREF

	err := wasmtime_table_grow(t.storeCtx, &t.val, delta, &val, &prevSize)
	if err != 0 {
		return 0, false
	}
	return prevSize, true
}

func (t *table) Get(ctx context.Context, index uint32) uint64 {
	var val wasmtime_val_t
	ok := wasmtime_table_get(t.storeCtx, &t.val, index, &val)
	if !ok {
		return 0
	}

	// Convert to uint64 based on type
	switch val.kind {
	case WASM_FUNCREF:
		funcRef := val.GetFuncRef()
		// Return the function's internal ID
		return funcRef.store_id
	case WASM_EXTERNREF:
		return uint64(val.GetExternRef())
	default:
		return 0
	}
}

func (t *table) Set(ctx context.Context, index uint32, v uint64) error {
	var val wasmtime_val_t
	// Wasmtime C API limitation: Cannot query table element type.
	// Assuming funcref as the default. For externref tables, this may not work correctly.
	val.kind = WASM_FUNCREF

	var funcRef wasmtime_func_t
	funcRef.store_id = v

	val.SetFuncRef(funcRef)

	ok := wasmtime_table_set(t.storeCtx, &t.val, index, &val)
	if !ok {
		return fmt.Errorf("failed to set table element at index %d", index)
	}
	return nil
}
