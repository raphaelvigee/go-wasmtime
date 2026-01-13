package wasmtime

import (
	"fmt"
	"runtime"
)

// Linker represents a wasmtime linker for resolving imports
type Linker struct {
	ptr    wasmtime_linker_t
	engine *Engine
}

// NewLinker creates a new linker for the given engine
func NewLinker(engine *Engine) (*Linker, error) {
	if err := Initialize(); err != nil {
		return nil, err
	}

	ptr := wasmtime_linker_new(engine.ptr)
	if ptr == 0 {
		return nil, fmt.Errorf("failed to create linker")
	}

	linker := &Linker{
		ptr:    ptr,
		engine: engine,
	}

	runtime.SetFinalizer(linker, func(l *Linker) {
		l.Close()
	})

	return linker, nil
}

// DefineWASI defines all WASI functions in this linker
func (l *Linker) DefineWASI() error {
	err := wasmtime_linker_define_wasi(l.ptr)
	if err != 0 {
		return fmt.Errorf("failed to define WASI: %w", getErrorMessage(err, 0))
	}
	return nil
}

// Instantiate instantiates a module using the linker to resolve imports
func (l *Linker) Instantiate(store *Store, module *Module) (*Instance, error) {
	var inst wasmtime_instance_t
	var trap *wasm_trap_t

	err := wasmtime_linker_instantiate(l.ptr, store.Context(), module.ptr, &inst, &trap)

	// Keep objects alive during C call
	runtime.KeepAlive(l)
	runtime.KeepAlive(store)
	runtime.KeepAlive(module)

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

// Close frees the linker resources
func (l *Linker) Close() {
	if l.ptr != 0 {
		wasmtime_linker_delete(l.ptr)
		l.ptr = 0
	}
}
