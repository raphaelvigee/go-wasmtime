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
	apply(storeCtx wasmtime_context_t, bindings *bindings) error
}

type wasiConfig struct {
	args          []string
	env           map[string]string
	preopenDirs   map[string]string // host -> guest
	inheritArgs   bool
	inheritEnv    bool
	inheritStdin  bool
	inheritStdout bool
	inheritStderr bool
}

// NewWASIConfig creates a new WASI configuration.
func NewWASIConfig() WASIConfig {
	return &wasiConfig{
		env:         make(map[string]string),
		preopenDirs: make(map[string]string),
	}
}

// WithArgs sets command-line arguments (variadic for easier use).
func (w *wasiConfig) WithArgs(args ...string) WASIConfig {
	w.args = args
	return w
}

// WithEnv sets a single environment variable.
func (w *wasiConfig) WithEnv(key, value string) WASIConfig {
	w.env[key] = value
	return w
}

// WithEnvs sets multiple environment variables.
func (w *wasiConfig) WithEnvs(env map[string]string) WASIConfig {
	for k, v := range env {
		w.env[k] = v
	}
	return w
}

// WithPreopenDir grants WASI access to a directory.
func (w *wasiConfig) WithPreopenDir(hostPath, guestPath string) WASIConfig {
	w.preopenDirs[hostPath] = guestPath
	return w
}

// WithInheritArgs inherits arguments from the host process.
func (w *wasiConfig) WithInheritArgs() WASIConfig {
	w.inheritArgs = true
	return w
}

// WithInheritEnv inherits environment variables from the host process.
func (w *wasiConfig) WithInheritEnv() WASIConfig {
	w.inheritEnv = true
	return w
}

// WithInheritStdin inherits stdin from the host process.
func (w *wasiConfig) WithInheritStdin() WASIConfig {
	w.inheritStdin = true
	return w
}

// WithInheritStdout inherits stdout from the host process.
func (w *wasiConfig) WithInheritStdout() WASIConfig {
	w.inheritStdout = true
	return w
}

// WithInheritStderr inherits stderr from the host process.
func (w *wasiConfig) WithInheritStderr() WASIConfig {
	w.inheritStderr = true
	return w
}

// WithInheritStdio inherits all stdio from the host process.
func (w *wasiConfig) WithInheritStdio() WASIConfig {
	return w.WithInheritStdin().WithInheritStdout().WithInheritStderr()
}

// apply applies the WASI configuration to a store context (internal method).
func (w *wasiConfig) apply(storeCtx wasmtime_context_t, bindings *bindings) error {
	// Create WASI config
	ptr := bindings.wasi_config_new()
	if ptr == 0 {
		return fmt.Errorf("failed to create WASI config")
	}

	// Apply arguments
	if w.inheritArgs {
		bindings.wasi_config_inherit_argv(ptr)
	} else if len(w.args) > 0 {
		argv := cStringArray(w.args)
		bindings.wasi_config_set_argv(ptr, int32(len(w.args)), argv)
	}

	// Apply environment
	if w.inheritEnv {
		bindings.wasi_config_inherit_env(ptr)
	} else if len(w.env) > 0 {
		names := make([]string, 0, len(w.env))
		values := make([]string, 0, len(w.env))
		for k, v := range w.env {
			names = append(names, k)
			values = append(values, v)
		}
		namesArr := cStringArray(names)
		valuesArr := cStringArray(values)
		bindings.wasi_config_set_env(ptr, int32(len(w.env)), namesArr, valuesArr)
	}

	// Apply preopen directories
	for hostPath, guestPath := range w.preopenDirs {
		bindings.wasi_config_preopen_dir(ptr, cString(hostPath), cString(guestPath))
	}

	// Apply stdio inheritance
	if w.inheritStdin {
		bindings.wasi_config_inherit_stdin(ptr)
	}
	if w.inheritStdout {
		bindings.wasi_config_inherit_stdout(ptr)
	}
	if w.inheritStderr {
		bindings.wasi_config_inherit_stderr(ptr)
	}

	// Set WASI context
	err := bindings.wasmtime_context_set_wasi(storeCtx, ptr)
	if err != 0 {
		return fmt.Errorf("failed to set WASI: %w", bindings.getErrorMessage(err, 0))
	}
	// Note: wasmtime takes ownership of the config, so we don't delete it
	return nil
}

// WASIExitError represents a WASI program exit.
type WASIExitError struct {
	ExitCode int32
}

func (e *WASIExitError) Error() string {
	return fmt.Sprintf("WASI program exited with status %d", e.ExitCode)
}
