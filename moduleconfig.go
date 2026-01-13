package wasmtime

import (
	"io"
	"io/fs"
)

// ModuleConfig configures a WebAssembly module instance.
// This matches wazero's ModuleConfig interface for compatibility.
type ModuleConfig interface {
	// WithName sets the name of the module.
	// This is used for error messages and debugging.
	WithName(name string) ModuleConfig

	// WithArgs sets command-line arguments for WASI modules.
	// The first argument is conventionally the program name.
	WithArgs(args ...string) ModuleConfig

	// WithEnv sets a single environment variable.
	WithEnv(key, value string) ModuleConfig

	// WithEnvs sets multiple environment variables from a map.
	WithEnvs(env map[string]string) ModuleConfig

	// WithStdin configures standard input.
	WithStdin(r io.Reader) ModuleConfig

	// WithStdout configures standard output.
	WithStdout(w io.Writer) ModuleConfig

	// WithStderr configures standard error.
	WithStderr(w io.Writer) ModuleConfig

	// WithFS sets the filesystem for WASI preopened directories.
	WithFS(filesystem fs.FS) ModuleConfig

	// WithDirPreopen grants access to a host directory.
	// The guestPath is the path within the WASM module, hostPath is the actual directory.
	WithDirPreopen(hostPath, guestPath string) ModuleConfig

	// WithStartFunctions controls which start functions to call.
	// An empty list means call the default _start function.
	WithStartFunctions(names ...string) ModuleConfig
}

// moduleConfig implements ModuleConfig
type moduleConfig struct {
	name           string
	args           []string
	env            map[string]string
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	filesystem     fs.FS
	preopens       map[string]string // guest -> host
	startFunctions []string
}

// NewModuleConfig creates a new module configuration with defaults.
func NewModuleConfig() ModuleConfig {
	return &moduleConfig{
		env:      make(map[string]string),
		preopens: make(map[string]string),
	}
}

func (mc *moduleConfig) WithName(name string) ModuleConfig {
	mc.name = name
	return mc
}

func (mc *moduleConfig) WithArgs(args ...string) ModuleConfig {
	mc.args = args
	return mc
}

func (mc *moduleConfig) WithEnv(key, value string) ModuleConfig {
	if mc.env == nil {
		mc.env = make(map[string]string)
	}
	mc.env[key] = value
	return mc
}

func (mc *moduleConfig) WithEnvs(env map[string]string) ModuleConfig {
	if mc.env == nil {
		mc.env = make(map[string]string)
	}
	for k, v := range env {
		mc.env[k] = v
	}
	return mc
}

func (mc *moduleConfig) WithStdin(r io.Reader) ModuleConfig {
	mc.stdin = r
	return mc
}

func (mc *moduleConfig) WithStdout(w io.Writer) ModuleConfig {
	mc.stdout = w
	return mc
}

func (mc *moduleConfig) WithStderr(w io.Writer) ModuleConfig {
	mc.stderr = w
	return mc
}

func (mc *moduleConfig) WithFS(filesystem fs.FS) ModuleConfig {
	mc.filesystem = filesystem
	return mc
}

func (mc *moduleConfig) WithDirPreopen(hostPath, guestPath string) ModuleConfig {
	if mc.preopens == nil {
		mc.preopens = make(map[string]string)
	}
	mc.preopens[guestPath] = hostPath
	return mc
}

func (mc *moduleConfig) WithStartFunctions(names ...string) ModuleConfig {
	mc.startFunctions = names
	return mc
}
