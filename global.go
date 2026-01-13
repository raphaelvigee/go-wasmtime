package wasmtime

import (
	"context"
	"fmt"

	"github.com/rvigee/purego-wasmtime/api"
)

// global implements api.Global for a WebAssembly global variable.
type global struct {
	val      wasmtime_global_t
	store    wasmtime_store_t
	storeCtx wasmtime_context_t
	valType  api.ValueType
	mutable  bool
}

func (g *global) Type() api.ValueType {
	return g.valType
}

func (g *global) Get(ctx context.Context) uint64 {
	var val wasmtime_val_t
	wasmtime_global_get(g.storeCtx, &g.val, &val)

	// Convert wasmtime_val_t to uint64
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

func (g *global) Set(ctx context.Context, v uint64) error {
	if !g.mutable {
		return fmt.Errorf("global is immutable")
	}

	var val wasmtime_val_t
	val.kind = uint8(g.valType)

	// Convert uint64 to wasmtime_val_t based on type
	switch g.valType {
	case api.ValueTypeI32:
		val.SetI32(int32(uint32(v)))
	case api.ValueTypeI64:
		val.SetI64(int64(v))
	case api.ValueTypeF32:
		val.SetF32(float32(v))
	case api.ValueTypeF64:
		val.SetF64(float64(v))
	case api.ValueTypeExternref:
		val.SetExternRef(uintptr(v))
	default:
		return fmt.Errorf("unsupported global type: %v", g.valType)
	}

	wasmtime_global_set(g.storeCtx, &g.val, &val)
	return nil
}
