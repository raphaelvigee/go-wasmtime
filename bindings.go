package wasmtime

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Library loading cache with reference counting
var (
	libCacheMu sync.Mutex
	libCache   = make(map[string]*libraryHandle)
)

type libraryHandle struct {
	handle   uintptr
	refCount int
}

// bindings holds all purego function pointers for a wasmtime instance
type bindings struct {
	// Engine functions
	wasm_engine_new    func() wasm_engine_t
	wasm_engine_delete func(wasm_engine_t)

	// Store functions
	wasmtime_store_new     func(wasm_engine_t, uintptr, uintptr) wasmtime_store_t
	wasmtime_store_delete  func(wasmtime_store_t)
	wasmtime_store_context func(wasmtime_store_t) wasmtime_context_t

	// WAT conversion
	wasmtime_wat2wasm func(*byte, uintptr, *wasm_byte_vec_t) wasmtime_error_t

	// Module functions
	wasmtime_module_new    func(wasm_engine_t, *byte, uintptr, *wasmtime_module_t) wasmtime_error_t
	wasmtime_module_delete func(wasmtime_module_t)

	// Instance functions
	wasmtime_instance_new        func(wasmtime_context_t, wasmtime_module_t, *wasmtime_extern_t, uintptr, *wasmtime_instance_t, **wasm_trap_t) wasmtime_error_t
	wasmtime_instance_export_get func(wasmtime_context_t, *wasmtime_instance_t, *byte, uintptr, *wasmtime_extern_t) bool

	// Function calling
	wasmtime_func_call func(wasmtime_context_t, *wasmtime_func_t, *wasmtime_val_t, uintptr, *wasmtime_val_t, uintptr, **wasm_trap_t) wasmtime_error_t
	wasmtime_func_type func(wasmtime_context_t, *wasmtime_func_t) wasm_functype_t

	// Error handling
	wasmtime_error_message     func(wasmtime_error_t, *wasm_byte_vec_t)
	wasmtime_error_delete      func(wasmtime_error_t)
	wasmtime_error_exit_status func(wasmtime_error_t, *int32) bool
	wasm_trap_new              func(*byte, uintptr) wasm_trap_t
	wasm_trap_message          func(wasm_trap_t, *wasm_byte_vec_t)
	wasm_trap_delete           func(wasm_trap_t)

	// Byte vectors
	wasm_byte_vec_new_uninitialized func(*wasm_byte_vec_t, uintptr)
	wasm_byte_vec_delete            func(*wasm_byte_vec_t)

	// Function type introspection
	wasm_functype_params  func(wasm_functype_t) *wasm_valtype_vec_t
	wasm_functype_results func(wasm_functype_t) *wasm_valtype_vec_t
	wasm_functype_delete  func(wasm_functype_t)
	wasm_valtype_kind     func(wasm_valtype_t) wasm_valkind_t

	// Linker functions
	wasmtime_linker_new         func(wasm_engine_t) wasmtime_linker_t
	wasmtime_linker_delete      func(wasmtime_linker_t)
	wasmtime_linker_define_wasi func(wasmtime_linker_t) wasmtime_error_t
	wasmtime_caller_export_get  func(uintptr, *byte, uintptr, *wasmtime_extern_t) bool
	wasmtime_linker_instantiate func(wasmtime_linker_t, wasmtime_context_t, wasmtime_module_t, *wasmtime_instance_t, **wasm_trap_t) wasmtime_error_t
	wasmtime_linker_define      func(wasmtime_linker_t, wasmtime_context_t, *byte, uintptr, *byte, uintptr, *wasmtime_extern_t) wasmtime_error_t

	// Memory functions
	wasmtime_memory_data      func(wasmtime_context_t, *wasmtime_memory_t) unsafe.Pointer
	wasmtime_memory_data_size func(wasmtime_context_t, *wasmtime_memory_t) uintptr
	wasmtime_memory_size      func(wasmtime_context_t, *wasmtime_memory_t) uint64
	wasmtime_memory_grow      func(wasmtime_context_t, *wasmtime_memory_t, uint64, *uint64) wasmtime_error_t

	// Global functions
	wasmtime_global_get func(wasmtime_context_t, *wasmtime_global_t, *wasmtime_val_t)
	wasmtime_global_set func(wasmtime_context_t, *wasmtime_global_t, *wasmtime_val_t)

	// Table functions
	wasmtime_table_size func(wasmtime_context_t, *wasmtime_table_t) uint32
	wasmtime_table_get  func(wasmtime_context_t, *wasmtime_table_t, uint32, *wasmtime_val_t) bool
	wasmtime_table_set  func(wasmtime_context_t, *wasmtime_table_t, uint32, *wasmtime_val_t) bool
	wasmtime_table_grow func(wasmtime_context_t, *wasmtime_table_t, uint32, *wasmtime_val_t, *uint32) wasmtime_error_t

	// Host function support
	wasmtime_func_new                  func(wasmtime_context_t, wasm_functype_t, uintptr, uintptr, uintptr, *wasmtime_func_t)
	wasm_functype_new                  func(*wasm_valtype_vec_t, *wasm_valtype_vec_t) wasm_functype_t
	wasm_valtype_new                   func(uint8) wasm_valtype_t
	wasm_valtype_delete                func(wasm_valtype_t)
	wasm_valtype_vec_new_empty         func(*wasm_valtype_vec_t)
	wasm_valtype_vec_new_uninitialized func(*wasm_valtype_vec_t, uintptr)
	wasm_valtype_vec_delete            func(*wasm_valtype_vec_t)

	// WASI bindings
	wasi_config_new            func() wasi_config_t
	wasi_config_delete         func(wasi_config_t)
	wasi_config_inherit_argv   func(wasi_config_t)
	wasi_config_inherit_env    func(wasi_config_t)
	wasi_config_set_argv       func(wasi_config_t, int32, **byte)
	wasi_config_set_env        func(wasi_config_t, int32, **byte, **byte)
	wasi_config_preopen_dir    func(wasi_config_t, *byte, *byte) bool
	wasi_config_inherit_stdin  func(wasi_config_t)
	wasi_config_inherit_stdout func(wasi_config_t)
	wasi_config_inherit_stderr func(wasi_config_t)
	wasmtime_context_set_wasi  func(wasmtime_context_t, wasi_config_t) wasmtime_error_t
}

// loadLibrary loads the wasmtime library with memoization and reference counting
func loadLibrary(path string) (uintptr, error) {
	libCacheMu.Lock()
	defer libCacheMu.Unlock()

	// Check if already loaded
	if handle, ok := libCache[path]; ok {
		handle.refCount++
		return handle.handle, nil
	}

	// Load the library
	libHandle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return 0, fmt.Errorf("failed to load wasmtime library: %w", err)
	}

	// Cache it
	libCache[path] = &libraryHandle{
		handle:   libHandle,
		refCount: 1,
	}

	return libHandle, nil
}

// releaseLibrary decrements the reference count and unloads the library if count reaches zero
func releaseLibrary(path string) {
	libCacheMu.Lock()
	defer libCacheMu.Unlock()

	handle, ok := libCache[path]
	if !ok {
		return // Already released or never loaded
	}

	handle.refCount--
	if handle.refCount <= 0 {
		// Unload the library
		purego.Dlclose(handle.handle)
		delete(libCache, path)
	}
}

// newBindings creates a new bindings instance and registers all functions
func newBindings(libHandle uintptr) (*bindings, error) {
	b := &bindings{}

	// Engine functions
	purego.RegisterLibFunc(&b.wasm_engine_new, libHandle, "wasm_engine_new")
	purego.RegisterLibFunc(&b.wasm_engine_delete, libHandle, "wasm_engine_delete")

	// Store functions
	purego.RegisterLibFunc(&b.wasmtime_store_new, libHandle, "wasmtime_store_new")
	purego.RegisterLibFunc(&b.wasmtime_store_delete, libHandle, "wasmtime_store_delete")
	purego.RegisterLibFunc(&b.wasmtime_store_context, libHandle, "wasmtime_store_context")

	// WAT conversion
	purego.RegisterLibFunc(&b.wasmtime_wat2wasm, libHandle, "wasmtime_wat2wasm")

	// Module functions
	purego.RegisterLibFunc(&b.wasmtime_module_new, libHandle, "wasmtime_module_new")
	purego.RegisterLibFunc(&b.wasmtime_module_delete, libHandle, "wasmtime_module_delete")

	// Instance functions
	purego.RegisterLibFunc(&b.wasmtime_instance_new, libHandle, "wasmtime_instance_new")
	purego.RegisterLibFunc(&b.wasmtime_instance_export_get, libHandle, "wasmtime_instance_export_get")

	// Function calling
	purego.RegisterLibFunc(&b.wasmtime_func_call, libHandle, "wasmtime_func_call")
	purego.RegisterLibFunc(&b.wasmtime_func_type, libHandle, "wasmtime_func_type")

	// Function type introspection
	purego.RegisterLibFunc(&b.wasm_functype_params, libHandle, "wasm_functype_params")
	purego.RegisterLibFunc(&b.wasm_functype_results, libHandle, "wasm_functype_results")
	purego.RegisterLibFunc(&b.wasm_functype_delete, libHandle, "wasm_functype_delete")
	purego.RegisterLibFunc(&b.wasm_valtype_kind, libHandle, "wasm_valtype_kind")

	// Error handling
	purego.RegisterLibFunc(&b.wasmtime_error_message, libHandle, "wasmtime_error_message")
	purego.RegisterLibFunc(&b.wasmtime_error_delete, libHandle, "wasmtime_error_delete")
	purego.RegisterLibFunc(&b.wasmtime_error_exit_status, libHandle, "wasmtime_error_exit_status")
	purego.RegisterLibFunc(&b.wasm_trap_new, libHandle, "wasm_trap_new")
	purego.RegisterLibFunc(&b.wasm_trap_message, libHandle, "wasm_trap_message")
	purego.RegisterLibFunc(&b.wasm_trap_delete, libHandle, "wasm_trap_delete")

	// Byte vectors
	purego.RegisterLibFunc(&b.wasm_byte_vec_new_uninitialized, libHandle, "wasm_byte_vec_new_uninitialized")
	purego.RegisterLibFunc(&b.wasm_byte_vec_delete, libHandle, "wasm_byte_vec_delete")

	// Linker functions
	purego.RegisterLibFunc(&b.wasmtime_linker_new, libHandle, "wasmtime_linker_new")
	purego.RegisterLibFunc(&b.wasmtime_linker_delete, libHandle, "wasmtime_linker_delete")
	purego.RegisterLibFunc(&b.wasmtime_linker_define_wasi, libHandle, "wasmtime_linker_define_wasi")
	purego.RegisterLibFunc(&b.wasmtime_linker_instantiate, libHandle, "wasmtime_linker_instantiate")
	purego.RegisterLibFunc(&b.wasmtime_linker_define, libHandle, "wasmtime_linker_define")

	// Memory functions
	purego.RegisterLibFunc(&b.wasmtime_memory_data, libHandle, "wasmtime_memory_data")
	purego.RegisterLibFunc(&b.wasmtime_memory_data_size, libHandle, "wasmtime_memory_data_size")
	purego.RegisterLibFunc(&b.wasmtime_memory_size, libHandle, "wasmtime_memory_size")
	purego.RegisterLibFunc(&b.wasmtime_memory_grow, libHandle, "wasmtime_memory_grow")

	// Global functions
	purego.RegisterLibFunc(&b.wasmtime_global_get, libHandle, "wasmtime_global_get")
	purego.RegisterLibFunc(&b.wasmtime_global_set, libHandle, "wasmtime_global_set")

	// Table functions
	purego.RegisterLibFunc(&b.wasmtime_table_size, libHandle, "wasmtime_table_size")
	purego.RegisterLibFunc(&b.wasmtime_table_get, libHandle, "wasmtime_table_get")
	purego.RegisterLibFunc(&b.wasmtime_table_set, libHandle, "wasmtime_table_set")
	purego.RegisterLibFunc(&b.wasmtime_table_grow, libHandle, "wasmtime_table_grow")

	// Host function support
	purego.RegisterLibFunc(&b.wasmtime_func_new, libHandle, "wasmtime_func_new")
	purego.RegisterLibFunc(&b.wasmtime_caller_export_get, libHandle, "wasmtime_caller_export_get")
	purego.RegisterLibFunc(&b.wasm_functype_new, libHandle, "wasm_functype_new")
	purego.RegisterLibFunc(&b.wasm_functype_delete, libHandle, "wasm_functype_delete")
	purego.RegisterLibFunc(&b.wasm_valtype_new, libHandle, "wasm_valtype_new")
	purego.RegisterLibFunc(&b.wasm_valtype_delete, libHandle, "wasm_valtype_delete")
	purego.RegisterLibFunc(&b.wasm_valtype_vec_new_empty, libHandle, "wasm_valtype_vec_new_empty")
	purego.RegisterLibFunc(&b.wasm_valtype_vec_new_uninitialized, libHandle, "wasm_valtype_vec_new_uninitialized")
	purego.RegisterLibFunc(&b.wasm_valtype_vec_delete, libHandle, "wasm_valtype_vec_delete")

	// WASI bindings
	purego.RegisterLibFunc(&b.wasi_config_new, libHandle, "wasi_config_new")
	purego.RegisterLibFunc(&b.wasi_config_delete, libHandle, "wasi_config_delete")
	purego.RegisterLibFunc(&b.wasi_config_inherit_argv, libHandle, "wasi_config_inherit_argv")
	purego.RegisterLibFunc(&b.wasi_config_inherit_env, libHandle, "wasi_config_inherit_env")
	purego.RegisterLibFunc(&b.wasi_config_set_argv, libHandle, "wasi_config_set_argv")
	purego.RegisterLibFunc(&b.wasi_config_set_env, libHandle, "wasi_config_set_env")
	purego.RegisterLibFunc(&b.wasi_config_preopen_dir, libHandle, "wasi_config_preopen_dir")
	purego.RegisterLibFunc(&b.wasi_config_inherit_stdin, libHandle, "wasi_config_inherit_stdin")
	purego.RegisterLibFunc(&b.wasi_config_inherit_stdout, libHandle, "wasi_config_inherit_stdout")
	purego.RegisterLibFunc(&b.wasi_config_inherit_stderr, libHandle, "wasi_config_inherit_stderr")
	purego.RegisterLibFunc(&b.wasmtime_context_set_wasi, libHandle, "wasmtime_context_set_wasi")

	return b, nil
}

// getErrorMessage extracts error message from wasmtime_error_t or wasm_trap_t
// Also detects WASI exits and returns WASIExitError for proper handling
func (b *bindings) getErrorMessage(err wasmtime_error_t, trap wasm_trap_t) error {
	if err == 0 && trap == 0 {
		return fmt.Errorf("unknown error")
	}

	// Check if this is a WASI exit before converting to string
	if err != 0 {
		var exitCode int32
		if b.wasmtime_error_exit_status(err, &exitCode) {
			// This is a WASI exit - return our typed error
			wasmtimeErr := &WASIExitError{ExitCode: exitCode}
			b.wasmtime_error_delete(err)
			return wasmtimeErr
		}
	}

	// Not a WASI exit, return regular error message
	var msg wasm_byte_vec_t
	if err != 0 {
		b.wasmtime_error_message(err, &msg)
		b.wasmtime_error_delete(err)
	} else if trap != 0 {
		b.wasm_trap_message(trap, &msg)
		b.wasm_trap_delete(trap)
	}

	result := string(msg.toGoBytes())
	b.wasm_byte_vec_delete(&msg)
	return fmt.Errorf("%s", result)
}
