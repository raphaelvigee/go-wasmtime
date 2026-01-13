package wasmtime

import (
	"fmt"
	"sync"

	"github.com/ebitengine/purego"
)

var (
	// Library handle
	libHandle uintptr

	// Function pointers - Engine
	wasm_engine_new    func() wasm_engine_t
	wasm_engine_delete func(wasm_engine_t)

	// Function pointers - Store
	wasmtime_store_new     func(wasm_engine_t, uintptr, uintptr) wasmtime_store_t
	wasmtime_store_delete  func(wasmtime_store_t)
	wasmtime_store_context func(wasmtime_store_t) wasmtime_context_t

	// Function pointers - WAT conversion
	wasmtime_wat2wasm func(*byte, uintptr, *wasm_byte_vec_t) wasmtime_error_t

	// Function pointers - Module
	wasmtime_module_new    func(wasm_engine_t, *byte, uintptr, *wasmtime_module_t) wasmtime_error_t
	wasmtime_module_delete func(wasmtime_module_t)

	// Function pointers - Instance
	wasmtime_instance_new        func(wasmtime_context_t, wasmtime_module_t, *wasmtime_extern_t, uintptr, *wasmtime_instance_t, **wasm_trap_t) wasmtime_error_t
	wasmtime_instance_export_get func(wasmtime_context_t, *wasmtime_instance_t, *byte, uintptr, *wasmtime_extern_t) bool

	// Function pointers - Function calling
	wasmtime_func_call func(wasmtime_context_t, *wasmtime_func_t, *wasmtime_val_t, uintptr, *wasmtime_val_t, uintptr, **wasm_trap_t) wasmtime_error_t
	wasmtime_func_type func(wasmtime_context_t, *wasmtime_func_t) wasm_functype_t

	// Function pointers - Error handling
	wasmtime_error_message func(wasmtime_error_t, *wasm_byte_vec_t)
	wasmtime_error_delete  func(wasmtime_error_t)
	wasm_trap_message      func(wasm_trap_t, *wasm_byte_vec_t)
	wasm_trap_delete       func(wasm_trap_t)

	// Function pointers - Byte vectors
	wasm_byte_vec_new_uninitialized func(*wasm_byte_vec_t, uintptr)
	wasm_byte_vec_delete            func(*wasm_byte_vec_t)

	// Function pointers - Function type introspection
	wasm_functype_results func(wasm_functype_t) *wasm_valtype_vec_t
	wasm_functype_delete  func(wasm_functype_t)

	initOnce sync.Once
	initErr  error
)

// Initialize loads the wasmtime library and binds all functions
func Initialize() error {
	initOnce.Do(func() {
		initErr = initializeImpl()
	})
	return initErr
}

func initializeImpl() error {
	// Get the library path (will auto-download if needed)
	libPath, err := getLibraryPath()
	if err != nil {
		return fmt.Errorf("failed to get wasmtime library: %w", err)
	}

	// Open the library
	libHandle, err = purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("failed to load wasmtime library: %w", err)
	}

	// Register all functions
	if err := registerFunctions(); err != nil {
		return fmt.Errorf("failed to register functions: %w", err)
	}

	return nil
}

func registerFunctions() error {
	// Engine functions
	purego.RegisterLibFunc(&wasm_engine_new, libHandle, "wasm_engine_new")
	purego.RegisterLibFunc(&wasm_engine_delete, libHandle, "wasm_engine_delete")

	// Store functions
	purego.RegisterLibFunc(&wasmtime_store_new, libHandle, "wasmtime_store_new")
	purego.RegisterLibFunc(&wasmtime_store_delete, libHandle, "wasmtime_store_delete")
	purego.RegisterLibFunc(&wasmtime_store_context, libHandle, "wasmtime_store_context")

	// WAT conversion
	purego.RegisterLibFunc(&wasmtime_wat2wasm, libHandle, "wasmtime_wat2wasm")

	// Module functions
	purego.RegisterLibFunc(&wasmtime_module_new, libHandle, "wasmtime_module_new")
	purego.RegisterLibFunc(&wasmtime_module_delete, libHandle, "wasmtime_module_delete")

	// Instance functions
	purego.RegisterLibFunc(&wasmtime_instance_new, libHandle, "wasmtime_instance_new")
	purego.RegisterLibFunc(&wasmtime_instance_export_get, libHandle, "wasmtime_instance_export_get")

	// Function calling
	purego.RegisterLibFunc(&wasmtime_func_call, libHandle, "wasmtime_func_call")
	purego.RegisterLibFunc(&wasmtime_func_type, libHandle, "wasmtime_func_type")

	// Function type introspection
	purego.RegisterLibFunc(&wasm_functype_results, libHandle, "wasm_functype_results")
	purego.RegisterLibFunc(&wasm_functype_delete, libHandle, "wasm_functype_delete")

	// Error handling
	purego.RegisterLibFunc(&wasmtime_error_message, libHandle, "wasmtime_error_message")
	purego.RegisterLibFunc(&wasmtime_error_delete, libHandle, "wasmtime_error_delete")
	purego.RegisterLibFunc(&wasm_trap_message, libHandle, "wasm_trap_message")
	purego.RegisterLibFunc(&wasm_trap_delete, libHandle, "wasm_trap_delete")

	// Byte vectors
	purego.RegisterLibFunc(&wasm_byte_vec_new_uninitialized, libHandle, "wasm_byte_vec_new_uninitialized")
	purego.RegisterLibFunc(&wasm_byte_vec_delete, libHandle, "wasm_byte_vec_delete")

	// Register WASI functions
	if err := registerWASIFunctions(); err != nil {
		return err
	}

	return nil
}

// getErrorMessage extracts error message from wasmtime_error_t or wasm_trap_t
func getErrorMessage(err wasmtime_error_t, trap wasm_trap_t) string {
	var msg wasm_byte_vec_t
	if err != 0 {
		wasmtime_error_message(err, &msg)
		wasmtime_error_delete(err)
	} else if trap != 0 {
		wasm_trap_message(trap, &msg)
		wasm_trap_delete(trap)
	} else {
		return "unknown error"
	}

	result := string(msg.toGoBytes())
	wasm_byte_vec_delete(&msg)
	return result
}
