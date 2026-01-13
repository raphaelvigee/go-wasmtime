package wasmtime

import (
	"github.com/ebitengine/purego"
)

var (
	// Function pointers - WASI Configuration
	wasi_config_new            func() wasi_config_t
	wasi_config_delete         func(wasi_config_t)
	wasi_config_set_argv       func(wasi_config_t, int, **byte) bool
	wasi_config_set_env        func(wasi_config_t, int, **byte, **byte) bool
	wasi_config_preopen_dir    func(wasi_config_t, *byte, *byte) bool
	wasi_config_inherit_argv   func(wasi_config_t)
	wasi_config_inherit_env    func(wasi_config_t)
	wasi_config_inherit_stdin  func(wasi_config_t)
	wasi_config_inherit_stdout func(wasi_config_t)
	wasi_config_inherit_stderr func(wasi_config_t)

	// Function pointers - WASI Context
	wasmtime_context_set_wasi func(wasmtime_context_t, wasi_config_t) wasmtime_error_t
)

func registerWASIFunctions() error {
	// WASI configuration
	purego.RegisterLibFunc(&wasi_config_new, libHandle, "wasi_config_new")
	purego.RegisterLibFunc(&wasi_config_delete, libHandle, "wasi_config_delete")
	purego.RegisterLibFunc(&wasi_config_set_argv, libHandle, "wasi_config_set_argv")
	purego.RegisterLibFunc(&wasi_config_set_env, libHandle, "wasi_config_set_env")
	purego.RegisterLibFunc(&wasi_config_preopen_dir, libHandle, "wasi_config_preopen_dir")
	purego.RegisterLibFunc(&wasi_config_inherit_argv, libHandle, "wasi_config_inherit_argv")
	purego.RegisterLibFunc(&wasi_config_inherit_env, libHandle, "wasi_config_inherit_env")
	purego.RegisterLibFunc(&wasi_config_inherit_stdin, libHandle, "wasi_config_inherit_stdin")
	purego.RegisterLibFunc(&wasi_config_inherit_stdout, libHandle, "wasi_config_inherit_stdout")
	purego.RegisterLibFunc(&wasi_config_inherit_stderr, libHandle, "wasi_config_inherit_stderr")

	// WASI context
	purego.RegisterLibFunc(&wasmtime_context_set_wasi, libHandle, "wasmtime_context_set_wasi")

	return nil
}
