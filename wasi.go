package wasmtime

import (
	"fmt"
)

// WASIConfig represents WASI configuration
type WASIConfig struct {
	ptr wasi_config_t
}

// NewWASIConfig creates a new WASI configuration
func NewWASIConfig() (*WASIConfig, error) {
	if err := Initialize(); err != nil {
		return nil, err
	}

	ptr := wasi_config_new()
	if ptr == 0 {
		return nil, fmt.Errorf("failed to create WASI config")
	}

	return &WASIConfig{ptr: ptr}, nil
}

// WithArgs sets command-line arguments
func (w *WASIConfig) WithArgs(args []string) *WASIConfig {
	if len(args) > 0 {
		argv := cStringArray(args)
		wasi_config_set_argv(w.ptr, len(args), argv)
	}
	return w
}

// WithEnv sets environment variables
func (w *WASIConfig) WithEnv(env map[string]string) *WASIConfig {
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

// WithPreopenDir grants WASI access to a directory
func (w *WASIConfig) WithPreopenDir(hostPath, guestPath string) *WASIConfig {
	wasi_config_preopen_dir(w.ptr, cString(hostPath), cString(guestPath))
	return w
}

// WithInheritArgs inherits arguments from the host process
func (w *WASIConfig) WithInheritArgs() *WASIConfig {
	wasi_config_inherit_argv(w.ptr)
	return w
}

// WithInheritEnv inherits environment variables from the host process
func (w *WASIConfig) WithInheritEnv() *WASIConfig {
	wasi_config_inherit_env(w.ptr)
	return w
}

// WithInheritStdin inherits stdin from the host process
func (w *WASIConfig) WithInheritStdin() *WASIConfig {
	wasi_config_inherit_stdin(w.ptr)
	return w
}

// WithInheritStdout inherits stdout from the host process
func (w *WASIConfig) WithInheritStdout() *WASIConfig {
	wasi_config_inherit_stdout(w.ptr)
	return w
}

// WithInheritStderr inherits stderr from the host process
func (w *WASIConfig) WithInheritStderr() *WASIConfig {
	wasi_config_inherit_stderr(w.ptr)
	return w
}

// WithInheritStdio inherits all stdio from the host process
func (w *WASIConfig) WithInheritStdio() *WASIConfig {
	return w.WithInheritStdin().WithInheritStdout().WithInheritStderr()
}

// Apply applies the WASI configuration to a store context
func (w *WASIConfig) Apply(store *Store) error {
	err := wasmtime_context_set_wasi(store.Context(), w.ptr)
	if err != 0 {
		return fmt.Errorf("failed to set WASI: %w", getErrorMessage(err, 0))
	}
	// Note: wasmtime takes ownership of the config, so we don't delete it
	w.ptr = 0
	return nil
}

// WASIExitError represents a WASI program exit
type WASIExitError struct {
	ExitCode int32
}

func (e *WASIExitError) Error() string {
	return fmt.Sprintf("WASI program exited with status %d", e.ExitCode)
}
