package wasmtime

import (
	"fmt"
	"runtime"
)

// Engine represents a WebAssembly compilation engine
type Engine struct {
	ptr wasm_engine_t
}

// NewEngine creates a new WebAssembly engine
func NewEngine() (*Engine, error) {
	if err := Initialize(); err != nil {
		return nil, err
	}

	ptr := wasm_engine_new()
	if ptr == 0 {
		return nil, fmt.Errorf("failed to create engine")
	}

	engine := &Engine{ptr: ptr}
	runtime.SetFinalizer(engine, (*Engine).Close)
	return engine, nil
}

// Close releases the engine resources
func (e *Engine) Close() {
	if e.ptr != 0 {
		wasm_engine_delete(e.ptr)
		e.ptr = 0
	}
}

// Store represents a WebAssembly store
type Store struct {
	ptr    wasmtime_store_t
	engine *Engine
}

// NewStore creates a new WebAssembly store
func NewStore(engine *Engine) (*Store, error) {
	ptr := wasmtime_store_new(engine.ptr, 0, 0)
	if ptr == 0 {
		return nil, fmt.Errorf("failed to create store")
	}

	store := &Store{ptr: ptr, engine: engine}
	runtime.SetFinalizer(store, (*Store).Close)
	return store, nil
}

// Context returns the context for this store
func (s *Store) Context() wasmtime_context_t {
	return wasmtime_store_context(s.ptr)
}

// Close releases the store resources
func (s *Store) Close() {
	if s.ptr != 0 {
		wasmtime_store_delete(s.ptr)
		s.ptr = 0
	}
}
