package wasmtime_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

// TestAllTypes comprehensive test for all WebAssembly value types
func TestAllTypes(t *testing.T) {
	ts := newTestSetup(t, "testdata/all_types.wat", false)
	defer ts.cleanup()

	// ========================================
	// i32 Tests
	// ========================================
	t.Run("i32_identity", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i32_identity")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), wasmtime.DecodeI32(results[0]))
	})

	t.Run("i32_add", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i32_add")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(10), wasmtime.EncodeI32(32))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), wasmtime.DecodeI32(results[0]))
	})

	t.Run("i32_multiply", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i32_multiply")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(6), wasmtime.EncodeI32(7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), wasmtime.DecodeI32(results[0]))
	})

	t.Run("i32_edge_cases", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i32_identity")
		require.NotNil(t, fn)

		// Test max value
		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(math.MaxInt32))
		require.NoError(t, err)
		assert.Equal(t, int32(math.MaxInt32), wasmtime.DecodeI32(results[0]))

		// Test min value
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI32(math.MinInt32))
		require.NoError(t, err)
		assert.Equal(t, int32(math.MinInt32), wasmtime.DecodeI32(results[0]))

		// Test zero
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI32(0))
		require.NoError(t, err)
		assert.Equal(t, int32(0), wasmtime.DecodeI32(results[0]))

		// Test negative
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI32(-42))
		require.NoError(t, err)
		assert.Equal(t, int32(-42), wasmtime.DecodeI32(results[0]))
	})

	// ========================================
	// i64 Tests
	// ========================================
	t.Run("i64_identity", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i64_identity")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI64(9223372036854775807))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int64(9223372036854775807), wasmtime.DecodeI64(results[0]))
	})

	t.Run("i64_add", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i64_add")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI64(100), wasmtime.EncodeI64(200))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int64(300), wasmtime.DecodeI64(results[0]))
	})

	t.Run("i64_multiply", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i64_multiply")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI64(6), wasmtime.EncodeI64(7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int64(42), wasmtime.DecodeI64(results[0]))
	})

	t.Run("i64_edge_cases", func(t *testing.T) {
		fn := ts.module.ExportedFunction("i64_identity")
		require.NotNil(t, fn)

		// Test max value
		results, err := fn.Call(ts.ctx, wasmtime.EncodeI64(math.MaxInt64))
		require.NoError(t, err)
		assert.Equal(t, int64(math.MaxInt64), wasmtime.DecodeI64(results[0]))

		// Test min value
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI64(math.MinInt64))
		require.NoError(t, err)
		assert.Equal(t, int64(math.MinInt64), wasmtime.DecodeI64(results[0]))

		// Test zero
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), wasmtime.DecodeI64(results[0]))
	})

	// ========================================
	// f32 Tests
	// ========================================
	t.Run("f32_identity", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f32_identity")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF32(3.14159))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(3.14159), wasmtime.DecodeF32(results[0]), 0.00001)
	})

	t.Run("f32_add", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f32_add")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF32(1.5), wasmtime.EncodeF32(2.5))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(4.0), wasmtime.DecodeF32(results[0]), 0.00001)
	})

	t.Run("f32_multiply", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f32_multiply")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF32(2.5), wasmtime.EncodeF32(4.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(10.0), wasmtime.DecodeF32(results[0]), 0.00001)
	})

	t.Run("f32_sqrt", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f32_sqrt")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF32(16.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(4.0), wasmtime.DecodeF32(results[0]), 0.00001)
	})

	// ========================================
	// f64 Tests
	// ========================================
	t.Run("f64_identity", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f64_identity")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF64(3.141592653589793))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(3.141592653589793), wasmtime.DecodeF64(results[0]))
	})

	t.Run("f64_add", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f64_add")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF64(1.5), wasmtime.EncodeF64(2.5))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(4.0), wasmtime.DecodeF64(results[0]))
	})

	t.Run("f64_multiply", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f64_multiply")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF64(2.5), wasmtime.EncodeF64(4.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(10.0), wasmtime.DecodeF64(results[0]))
	})

	t.Run("f64_sqrt", func(t *testing.T) {
		fn := ts.module.ExportedFunction("f64_sqrt")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeF64(16.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(4.0), wasmtime.DecodeF64(results[0]))
	})

	// ========================================
	// Mixed Type Tests
	// ========================================
	t.Run("mixed_params", func(t *testing.T) {
		fn := ts.module.ExportedFunction("mixed_params")
		require.NotNil(t, fn)

		// Test: (i32 + i64) * f32 + f64
		// (10 + 5) * 2.0 + 0.5 = 30.5
		results, err := fn.Call(ts.ctx,
			wasmtime.EncodeI32(10),
			wasmtime.EncodeI64(5),
			wasmtime.EncodeF32(2.0),
			wasmtime.EncodeF64(0.5))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float64(30.5), wasmtime.DecodeF64(results[0]), 0.001)
	})

	t.Run("multi_return", func(t *testing.T) {
		fn := ts.module.ExportedFunction("multi_return")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx)
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Check each return value type and value
		assert.Equal(t, int32(42), wasmtime.DecodeI32(results[0]))
		assert.InDelta(t, float64(3.14159), wasmtime.DecodeF64(results[1]), 0.00001)
		assert.Equal(t, int64(9223372036854775807), wasmtime.DecodeI64(results[2]))
	})
}
