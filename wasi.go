package wasmtime

import (
	"fmt"
)

// WASIConfig represents WASI configuration.
type WASIConfig interface {
	// WithArgs sets command-line arguments (variadic).
	WithArgs(args ...string) WASIConfig

	// WithEnv sets a single environment variable.
	WithEnv(key, value string) WASIConfig

	// WithEnvs sets multiple environment variables.
	WithEnvs(env map[string]string) WASIConfig

	// WithPreopenDir grants WASI access to a directory.
	WithPreopenDir(hostPath, guestPath string) WASIConfig

	// WithInheritArgs inherits arguments from the host process.
	WithInheritArgs() WASIConfig

	// WithInheritEnv inherits environment variables from the host process.
	WithInheritEnv() WASIConfig

	// WithInheritStdin inherits stdin from the host process.
	WithInheritStdin() WASIConfig

	// WithInheritStdout inherits stdout from the host process.
	WithInheritStdout() WASIConfig

	// WithInheritStderr inherits stderr from the host process.
	WithInheritStderr() WASIConfig

	// WithInheritStdio inherits all stdio from the host process.
	WithInheritStdio() WASIConfig

	// apply is an internal method to apply configuration to a store context
	apply(storeCtx wasmtime_context_t) error
}

type wasiConfig struct {
	ptr wasi_config_t
}

// NewWASIConfig creates a new WASI configuration.
func NewWASIConfig() WASIConfig {
	if err := Initialize(); err != nil {
		// Can't return error, so panic - should rarely happen as Initialize is safe to call multiple times
		panic(fmt.Sprintf("failed to initialize wasmtime: %v", err))
	}

	ptr := wasi_config_new()
	if ptr == 0 {
		panic("failed to create WASI config")
	}

	return &wasiConfig{ptr: ptr}
}

// WithArgs sets command-line arguments (variadic for easier use).
func (w *wasiConfig) WithArgs(args ...string) WASIConfig {
	if len(args) > 0 {
		argv := cStringArray(args)
		wasi_config_set_argv(w.ptr, len(args), argv)
	}
	return w
}

// WithEnv sets a single environment variable.
func (w *wasiConfig) WithEnv(key, value string) WASIConfig {
	names := []string{key}
	values := []string{value}
	namesArr := cStringArray(names)
	valuesArr := cStringArray(values)
	wasi_config_set_env(w.ptr, 1, namesArr, valuesArr)
	return w
}

// WithEnvs sets multiple environment variables.
func (w *wasiConfig) WithEnvs(env map[string]string) WASIConfig {
	if len(env) == 0 {
		return w
	}

	names := make([]string, 0, len(env))
	values := make([]string, 0, len(env))
	for k, v := range env {
		names = append(names, k)
		values = append(values, v)
	}

	namesArr := cStringArray(names)
	valuesArr := cStringArray(values)
	wasi_config_set_env(w.ptr, len(env), namesArr, valuesArr)

	return w
}

// WithPreopenDir grants WASI access to a directory.
func (w *wasiConfig) WithPreopenDir(hostPath, guestPath string) WASIConfig {
	wasi_config_preopen_dir(w.ptr, cString(hostPath), cString(guestPath))
	return w
}

// WithInheritArgs inherits arguments from the host process.
func (w *wasiConfig) WithInheritArgs() WASIConfig {
	wasi_config_inherit_argv(w.ptr)
	return w
}

// WithInheritEnv inherits environment variables from the host process.
func (w *wasiConfig) WithInheritEnv() WASIConfig {
	wasi_config_inherit_env(w.ptr)
	return w
}

// WithInheritStdin inherits stdin from the host process.
func (w *wasiConfig) WithInheritStdin() WASIConfig {
	wasi_config_inherit_stdin(w.ptr)
	return w
}

// WithInheritStdout inherits stdout from the host process.
func (w *wasiConfig) WithInheritStdout() WASIConfig {
	wasi_config_inherit_stdout(w.ptr)
	return w
}

// WithInheritStderr inherits stderr from the host process.
func (w *wasiConfig) WithInheritStderr() WASIConfig {
	wasi_config_inherit_stderr(w.ptr)
	return w
}

// WithInheritStdio inherits all stdio from the host process.
func (w *wasiConfig) WithInheritStdio() WASIConfig {
	return w.WithInheritStdin().WithInheritStdout().WithInheritStderr()
}

// apply applies the WASI configuration to a store context (internal method).
func (w *wasiConfig) apply(storeCtx wasmtime_context_t) error {
	err := wasmtime_context_set_wasi(storeCtx, w.ptr)
	if err != 0 {
		return fmt.Errorf("failed to set WASI: %w", getErrorMessage(err, 0))
	}
	// Note: wasmtime takes ownership of the config, so we don't delete it
	w.ptr = 0
	return nil
}

// WASIExitError represents a WASI program exit.
type WASIExitError struct {
	ExitCode int32
}

func (e *WASIExitError) Error() string {
	return fmt.Sprintf("WASI program exited with status %d", e.ExitCode)
}
