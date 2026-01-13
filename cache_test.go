package wasmtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompilationCache(t *testing.T) {
	t.Run("in_memory_cache", func(t *testing.T) {
		cache := NewCompilationCache()
		require.NotNil(t, cache)
		assert.NoError(t, cache.Close(t.Context()))
	})

	t.Run("disk_cache_creation", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "wasmtime-cache-test")
		defer os.RemoveAll(tmpDir)

		cache, err := NewCompilationCacheWithDir(tmpDir)
		require.NoError(t, err)

		// Verify directory was created
		_, err = os.Stat(tmpDir)
		assert.False(t, os.IsNotExist(err), "Cache directory was not created")

		assert.NoError(t, cache.Close(t.Context()))
	})

	t.Run("disk_cache_invalid_path", func(t *testing.T) {
		// Try to create cache at a file instead of directory
		tmpFile := filepath.Join(os.TempDir(), "wasmtime-cache-file-test")
		require.NoError(t, os.WriteFile(tmpFile, []byte("test"), 0644))
		defer os.Remove(tmpFile)

		_, err := NewCompilationCacheWithDir(tmpFile)
		assert.Error(t, err, "Expected error when cache path is a file")
	})

	t.Run("cache_with_runtime_config", func(t *testing.T) {
		cache := NewCompilationCache()
		config := NewRuntimeConfig().WithCompilationCache(cache)

		rc := config.(*runtimeConfig)
		assert.Equal(t, cache, rc.cache)

		cache.Close(t.Context())
	})

	t.Run("multiple_caches", func(t *testing.T) {
		cache1 := NewCompilationCache()
		cache2 := NewCompilationCache()

		// Verify they are different instances (different pointers)
		assert.NotSame(t, cache1, cache2, "NewCompilationCache should return different instances")

		cache1.Close(t.Context())
		cache2.Close(t.Context())
	})

}

func TestRuntimeWithCache(t *testing.T) {

	t.Run("runtime_with_cache", func(t *testing.T) {
		cache := NewCompilationCache()
		defer cache.Close(t.Context())

		config := NewRuntimeConfig().WithCompilationCache(cache)

		r, err := NewRuntimeWithConfig(t.Context(), config)
		require.NoError(t, err)
		defer r.Close(t.Context())

		// Compile a simple module
		wat := `(module (func (export "noop")))`

		compiled, err := r.CompileModule(t.Context(), []byte(wat))
		require.NoError(t, err)
		defer compiled.Close()
	})

	t.Run("runtime_without_cache", func(t *testing.T) {
		// Should work fine without cache
		config := NewRuntimeConfig()

		r, err := NewRuntimeWithConfig(t.Context(), config)
		require.NoError(t, err)
		defer r.Close(t.Context())

		wat := `(module (func (export "noop")))`

		compiled, err := r.CompileModule(t.Context(), []byte(wat))
		require.NoError(t, err)
		defer compiled.Close()
	})
}
