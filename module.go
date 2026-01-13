package wasmtime

import (
	"fmt"
	"os"
	"runtime"
)

// Module represents a compiled WebAssembly module
type Module struct {
	ptr    wasmtime_module_t
	engine *Engine
}

// NewModuleFromWAT creates a new module from WAT (WebAssembly Text) source
func NewModuleFromWAT(engine *Engine, wat string) (*Module, error) {
	// Convert WAT to WASM
	watBytes := []byte(wat)
	var wasmVec wasm_byte_vec_t

	err := wasmtime_wat2wasm(cString(wat), uintptr(len(watBytes)), &wasmVec)
	if err != 0 {
		return nil, fmt.Errorf("failed to parse WAT: %w", getErrorMessage(err, 0))
	}
	defer wasm_byte_vec_delete(&wasmVec)

	// Compile module
	return NewModuleFromWASM(engine, wasmVec.toGoBytes())
}

// NewModuleFromWATFile creates a new module from a WAT file
func NewModuleFromWATFile(engine *Engine, filename string) (*Module, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return NewModuleFromWAT(engine, string(data))
}

// NewModuleFromWASM creates a new module from WASM (WebAssembly binary) bytes
func NewModuleFromWASM(engine *Engine, wasm []byte) (*Module, error) {
	var modulePtr wasmtime_module_t

	wasmVec := newByteVec(wasm)
	err := wasmtime_module_new(engine.ptr, wasmVec.data, wasmVec.size, &modulePtr)
	if err != 0 {
		return nil, fmt.Errorf("failed to compile module: %w", getErrorMessage(err, 0))
	}

	module := &Module{ptr: modulePtr, engine: engine}
	runtime.SetFinalizer(module, (*Module).Close)
	return module, nil
}

// NewModuleFromWASMFile creates a new module from a WASM file
func NewModuleFromWASMFile(engine *Engine, filename string) (*Module, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return NewModuleFromWASM(engine, data)
}

// Close releases the module resources
func (m *Module) Close() {
	if m.ptr != 0 {
		wasmtime_module_delete(m.ptr)
		m.ptr = 0
	}
}
