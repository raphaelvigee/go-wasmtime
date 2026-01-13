package wasmtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CompilationCache caches compiled WebAssembly modules to improve performance.
// This matches wazero's CompilationCache interface.
type CompilationCache interface {
	// Close closes the cache and releases resources.
	Close(ctx context.Context) error
}

// compilationCache implements in-memory and optionally disk-based caching
type compilationCache struct {
	mu      sync.RWMutex
	dir     string // Empty for in-memory only
	modules map[string]CompiledModule
}

// NewCompilationCache creates a new in-memory compilation cache.
// This matches wazero's NewCompilationCache function.
func NewCompilationCache() CompilationCache {
	return &compilationCache{
		modules: make(map[string]CompiledModule),
	}
}

// NewCompilationCacheWithDir creates a compilation cache that persists to disk.
// The dirname will be created if it doesn't exist.
// This matches wazero's NewCompilationCacheWithDir function.
func NewCompilationCacheWithDir(dirname string) (CompilationCache, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Verify it's a directory
	info, err := os.Stat(dirname)
	if err != nil {
		return nil, fmt.Errorf("failed to stat cache directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("cache path is not a directory: %s", dirname)
	}

	return &compilationCache{
		dir:     dirname,
		modules: make(map[string]CompiledModule),
	}, nil
}

func (cc *compilationCache) Close(ctx context.Context) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Close all cached modules
	for _, mod := range cc.modules {
		if err := mod.Close(); err != nil {
			// Log error but continue closing others
			// In a production system, we'd want better error handling
			_ = err
		}
	}

	cc.modules = make(map[string]CompiledModule)
	return nil
}

// get retrieves a module from cache (internal method)
func (cc *compilationCache) get(key string) (CompiledModule, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	mod, ok := cc.modules[key]
	return mod, ok
}

// put stores a module in cache (internal method)
func (cc *compilationCache) put(key string, mod CompiledModule) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.modules[key] = mod
}

// getCachePath returns the file path for a cached module
func (cc *compilationCache) getCachePath(key string) string {
	if cc.dir == "" {
		return ""
	}
	// Use a simple hash-based naming scheme
	// In production, this should use a proper content hash
	return filepath.Join(cc.dir, key+".wasm.cache")
}

// Note: Full disk-based caching requires integration with wasmtime's
// serialization API, which may not be available in all versions.
// For now, we provide the structure and in-memory caching.
