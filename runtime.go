package wasmtime

import (
	"context"
	"fmt"
	"runtime"

	"github.com/rvigee/purego-wasmtime/api"
)

// Runtime is a WebAssembly runtime that can compile and instantiate modules.
type Runtime interface {
	// CompileModule compiles WebAssembly binary (WAT or WASM) into a CompiledModule.
	CompileModule(ctx context.Context, binary []byte) (CompiledModule, error)

	// Instantiate instantiates a compiled module without WASI.
	Instantiate(ctx context.Context, compiled CompiledModule) (api.Module, error)

	// InstantiateWithWASI instantiates a compiled module with WASI support.
	InstantiateWithWASI(ctx context.Context, compiled CompiledModule) (api.Module, error)

	// NewHostModuleBuilder creates a builder for defining host modules (Go functions).
	NewHostModuleBuilder(name string) HostModuleBuilder

	// Close closes the runtime and releases resources.
	Close(ctx context.Context) error
}

// RuntimeConfig configures the runtime behavior.
type RuntimeConfig interface {
	// WithWASI configures WASI for this runtime.
	WithWASI(WASIConfig) RuntimeConfig

	// WithCompilationCache sets the compilation cache for this runtime.
	WithCompilationCache(cache CompilationCache) RuntimeConfig
}

type runtimeConfig struct {
	wasiConfig   WASIConfig
	cache        CompilationCache
	libraryPath  string
	autoDownload bool
	version      string
}

func (rc *runtimeConfig) WithWASI(wasi WASIConfig) RuntimeConfig {
	rc.wasiConfig = wasi
	return rc
}

func (rc *runtimeConfig) WithCompilationCache(cache CompilationCache) RuntimeConfig {
	rc.cache = cache
	return rc
}

func (rc *runtimeConfig) WithLibraryPath(path string) RuntimeConfig {
	rc.libraryPath = path
	rc.autoDownload = false // Disable auto-download when custom path is set
	return rc
}

func (rc *runtimeConfig) WithAutoDownload(version string) RuntimeConfig {
	rc.autoDownload = true
	rc.version = version
	return rc
}

// NewRuntimeConfig creates a new runtime configuration.
func NewRuntimeConfig() RuntimeConfig {
	return &runtimeConfig{
		autoDownload: true,            // Default to auto-download
		version:      wasmtimeVersion, // Default version
	}
}

type wasmRuntime struct {
	engine      wasm_engine_t
	store       wasmtime_store_t
	linker      wasmtime_linker_t
	config      *runtimeConfig
	bindings    *bindings
	libraryPath string
}

// NewRuntime creates a new WebAssembly runtime with default configuration.
func NewRuntime(ctx context.Context) (Runtime, error) {
	return NewRuntimeWithConfig(ctx, NewRuntimeConfig())
}

// NewRuntimeWithConfig creates a new WebAssembly runtime with the given configuration.
func NewRuntimeWithConfig(ctx context.Context, config RuntimeConfig) (Runtime, error) {
	rc, ok := config.(*runtimeConfig)
	if !ok {
		rc = &runtimeConfig{
			autoDownload: true,
		}
	}

	// Determine library path
	var libPath string
	var err error
	if rc.libraryPath != "" {
		// Use custom library path
		libPath = rc.libraryPath
	} else {
		// Get library from cache or auto-download
		libPath, err = getLibraryPath(rc.autoDownload, rc.version)
		if err != nil {
			return nil, fmt.Errorf("failed to get wasmtime library: %w", err)
		}
	}

	// Load library with memoization
	libHandle, err := loadLibrary(libPath)
	if err != nil {
		return nil, err
	}

	// Create bindings
	bindings, err := newBindings(libHandle)
	if err != nil {
		releaseLibrary(libPath)
		return nil, fmt.Errorf("failed to create bindings: %w", err)
	}

	//Create engine
	enginePtr := bindings.wasm_engine_new()
	if enginePtr == 0 {
		releaseLibrary(libPath)
		return nil, fmt.Errorf("failed to create engine")
	}

	// Create store
	storePtr := bindings.wasmtime_store_new(enginePtr, 0, 0)
	if storePtr == 0 {
		bindings.wasm_engine_delete(enginePtr)
		releaseLibrary(libPath)
		return nil, fmt.Errorf("failed to create store")
	}

	// Create linker
	linkerPtr := bindings.wasmtime_linker_new(enginePtr)
	if linkerPtr == 0 {
		bindings.wasmtime_store_delete(storePtr)
		bindings.wasm_engine_delete(enginePtr)
		releaseLibrary(libPath)
		return nil, fmt.Errorf("failed to create linker")
	}

	r := &wasmRuntime{
		engine:      enginePtr,
		store:       storePtr,
		linker:      linkerPtr,
		config:      rc,
		bindings:    bindings,
		libraryPath: libPath,
	}

	runtime.SetFinalizer(r, (*wasmRuntime).finalize)

	return r, nil
}

func (r *wasmRuntime) CompileModule(ctx context.Context, binary []byte) (CompiledModule, error) {
	// Try to compile as WASM first
	var modulePtr wasmtime_module_t
	wasmVec := newByteVec(binary)

	err := r.bindings.wasmtime_module_new(r.engine, wasmVec.data, wasmVec.size, &modulePtr)
	if err == 0 {
		// Successfully compiled as WASM
		return &compiledModule{
			ptr:      modulePtr,
			engine:   r.engine,
			bindings: r.bindings,
		}, nil
	}

	// Try to parse as WAT and convert to WASM
	var wasmVec2 wasm_byte_vec_t
	watErr := r.bindings.wasmtime_wat2wasm(cString(string(binary)), uintptr(len(binary)), &wasmVec2)
	if watErr != 0 {
		return nil, fmt.Errorf("failed to compile module: not valid WASM or WAT: %w", r.bindings.getErrorMessage(err, 0))
	}
	defer r.bindings.wasm_byte_vec_delete(&wasmVec2)

	// Compile the converted WASM
	wasmBytes := wasmVec2.toGoBytes()
	wasmVec3 := newByteVec(wasmBytes)
	err2 := r.bindings.wasmtime_module_new(r.engine, wasmVec3.data, wasmVec3.size, &modulePtr)
	if err2 != 0 {
		return nil, fmt.Errorf("failed to compile module from WAT: %w", r.bindings.getErrorMessage(err2, 0))
	}

	return &compiledModule{
		ptr:      modulePtr,
		engine:   r.engine,
		bindings: r.bindings,
	}, nil
}

func (r *wasmRuntime) Instantiate(ctx context.Context, compiled CompiledModule) (api.Module, error) {
	cm, ok := compiled.(*compiledModule)
	if !ok {
		return nil, fmt.Errorf("invalid compiled module type")
	}

	var inst wasmtime_instance_t
	var trap *wasm_trap_t

	storeCtx := r.bindings.wasmtime_store_context(r.store)

	// Use linker to instantiate - this allows host functions to be resolved
	// This matches wazero's behavior where the runtime automatically links imports
	err := r.bindings.wasmtime_linker_instantiate(r.linker, storeCtx, cm.ptr, &inst, &trap)

	runtime.KeepAlive(r)
	runtime.KeepAlive(cm)

	if err != 0 {
		return nil, fmt.Errorf("failed to instantiate: %w", r.bindings.getErrorMessage(err, 0))
	}
	if trap != nil {
		return nil, fmt.Errorf("failed to instantiate (trap): %w", r.bindings.getErrorMessage(0, *trap))
	}

	return &module{
		inst:     inst,
		store:    r.store,
		bindings: r.bindings,
	}, nil
}

func (r *wasmRuntime) InstantiateWithWASI(ctx context.Context, compiled CompiledModule) (api.Module, error) {
	cm, ok := compiled.(*compiledModule)
	if !ok {
		return nil, fmt.Errorf("invalid compiled module type")
	}

	// Apply WASI configuration if provided
	if r.config.wasiConfig != nil {
		storeCtx := r.bindings.wasmtime_store_context(r.store)
		if err := r.config.wasiConfig.apply(storeCtx, r.bindings); err != nil {
			return nil, fmt.Errorf("failed to apply WASI config: %w", err)
		}
	}

	// Define WASI in linker
	err := r.bindings.wasmtime_linker_define_wasi(r.linker)
	if err != 0 {
		return nil, fmt.Errorf("failed to define WASI: %w", r.bindings.getErrorMessage(err, 0))
	}

	// Instantiate using linker
	var inst wasmtime_instance_t
	var trap *wasm_trap_t

	storeCtx := r.bindings.wasmtime_store_context(r.store)
	err2 := r.bindings.wasmtime_linker_instantiate(r.linker, storeCtx, cm.ptr, &inst, &trap)

	runtime.KeepAlive(r)
	runtime.KeepAlive(cm)

	if err2 != 0 {
		return nil, fmt.Errorf("failed to instantiate with WASI: %w", r.bindings.getErrorMessage(err2, 0))
	}
	if trap != nil {
		return nil, fmt.Errorf("failed to instantiate with WASI (trap): %w", r.bindings.getErrorMessage(0, *trap))
	}

	return &module{
		inst:     inst,
		store:    r.store,
		bindings: r.bindings,
	}, nil
}

func (r *wasmRuntime) Close(ctx context.Context) error {
	runtime.SetFinalizer(r, nil) // Prevent finalizer from running since we are closing explicitly
	r.finalize()
	return nil
}

func (r *wasmRuntime) finalize() {
	if r.linker != 0 {
		r.bindings.wasmtime_linker_delete(r.linker)
		r.linker = 0
	}
	if r.store != 0 {
		r.bindings.wasmtime_store_delete(r.store)
		r.store = 0
	}
	if r.engine != 0 {
		r.bindings.wasm_engine_delete(r.engine)
		r.engine = 0
	}
	// Release library reference
	if r.libraryPath != "" {
		releaseLibrary(r.libraryPath)
	}
}

// CompiledModule represents a compiled WebAssembly module.
type CompiledModule interface {
	// Close releases the compiled module.
	Close() error

	// Name returns the module name encoded in the binary, or empty if not set.
	Name() string

	// ImportedFunctions returns all imported functions or nil if there are none.
	ImportedFunctions() []api.FunctionDefinition

	// ExportedFunctions returns all exported functions keyed by export name.
	ExportedFunctions() map[string]api.FunctionDefinition

	// ImportedMemories returns all imported memories or nil if there are none.
	ImportedMemories() []api.MemoryDefinition

	// ExportedMemories returns all exported memories keyed by export name.
	ExportedMemories() map[string]api.MemoryDefinition

	// CustomSections returns all custom sections keyed by section name.
	CustomSections() []api.CustomSection
}

type compiledModule struct {
	ptr      wasmtime_module_t
	engine   wasm_engine_t
	bindings *bindings
}

func (cm *compiledModule) Close() error {
	if cm.ptr != 0 {
		cm.bindings.wasmtime_module_delete(cm.ptr)
		cm.ptr = 0
	}
	return nil
}

func (cm *compiledModule) Name() string {
	// Wasmtime C API limitation: wasmtime_module_name function is not available.
	// Module introspection must be done via parsing the WASM binary or after instantiation.
	return ""
}

func (cm *compiledModule) ImportedFunctions() []api.FunctionDefinition {
	// Wasmtime C API limitation: Module import/export introspection APIs are not available.
	// The C API does not expose functions like wasmtime_module_imports().
	// Workaround: Use the module's exports after instantiation via ExportedFunction().
	return nil
}

func (cm *compiledModule) ExportedFunctions() map[string]api.FunctionDefinition {
	// Wasmtime C API limitation: Module export introspection APIs are not available.
	// The C API does not expose functions like wasmtime_module_exports().
	// Workaround: Use the instantiated module's ExportedFunction() method.
	return make(map[string]api.FunctionDefinition)
}

func (cm *compiledModule) ImportedMemories() []api.MemoryDefinition {
	// Wasmtime C API limitation: Module import introspection APIs are not available.
	return nil
}

func (cm *compiledModule) ExportedMemories() map[string]api.MemoryDefinition {
	// Wasmtime C API limitation: Module export introspection APIs are not available.
	// Workaround: Use the instantiated module's ExportedMemory() method.
	return make(map[string]api.MemoryDefinition)
}

func (cm *compiledModule) CustomSections() []api.CustomSection {
	// Wasmtime C API limitation: Custom section APIs are not exposed in the C API.
	// These would require parsing the WASM binary directly.
	return nil
}
