package wasmtime

import (
	"github.com/ebitengine/purego"
)

var (
	// Function pointers - Linker
	wasmtime_linker_new         func(wasm_engine_t) wasmtime_linker_t
	wasmtime_linker_delete      func(wasmtime_linker_t)
	wasmtime_linker_define_wasi func(wasmtime_linker_t) wasmtime_error_t
	wasmtime_linker_instantiate func(wasmtime_linker_t, wasmtime_context_t, wasmtime_module_t, *wasmtime_instance_t, **wasm_trap_t) wasmtime_error_t
)

func registerLinkerFunctions() error {
	// Linker functions
	purego.RegisterLibFunc(&wasmtime_linker_new, libHandle, "wasmtime_linker_new")
	purego.RegisterLibFunc(&wasmtime_linker_delete, libHandle, "wasmtime_linker_delete")
	purego.RegisterLibFunc(&wasmtime_linker_define_wasi, libHandle, "wasmtime_linker_define_wasi")
	purego.RegisterLibFunc(&wasmtime_linker_instantiate, libHandle, "wasmtime_linker_instantiate")

	return nil
}
