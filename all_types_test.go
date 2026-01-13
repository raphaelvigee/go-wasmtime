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
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/all_types.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	// ========================================
	// i32 Tests
	// ========================================
	t.Run("i32_identity", func(t *testing.T) {
		results, err := instance.Call("i32_identity", int32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), results[0])
	})

	t.Run("i32_add", func(t *testing.T) {
		results, err := instance.Call("i32_add", int32(10), int32(32))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), results[0])
	})

	t.Run("i32_multiply", func(t *testing.T) {
		results, err := instance.Call("i32_multiply", int32(6), int32(7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), results[0])
	})

	t.Run("i32_edge_cases", func(t *testing.T) {
		// Test max value
		results, err := instance.Call("i32_identity", int32(math.MaxInt32))
		require.NoError(t, err)
		assert.Equal(t, int32(math.MaxInt32), results[0])

		// Test min value
		results, err = instance.Call("i32_identity", int32(math.MinInt32))
		require.NoError(t, err)
		assert.Equal(t, int32(math.MinInt32), results[0])

		// Test zero
		results, err = instance.Call("i32_identity", int32(0))
		require.NoError(t, err)
		assert.Equal(t, int32(0), results[0])

		// Test negative
		results, err = instance.Call("i32_identity", int32(-42))
		require.NoError(t, err)
		assert.Equal(t, int32(-42), results[0])
	})

	// ========================================
	// i64 Tests
	// ========================================
	t.Run("i64_identity", func(t *testing.T) {
		results, err := instance.Call("i64_identity", int64(9223372036854775807))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int64(9223372036854775807), results[0])
	})

	t.Run("i64_add", func(t *testing.T) {
		results, err := instance.Call("i64_add", int64(100), int64(200))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int64(300), results[0])
	})

	t.Run("i64_multiply", func(t *testing.T) {
		results, err := instance.Call("i64_multiply", int64(6), int64(7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int64(42), results[0])
	})

	t.Run("i64_edge_cases", func(t *testing.T) {
		// Test max value
		results, err := instance.Call("i64_identity", int64(math.MaxInt64))
		require.NoError(t, err)
		assert.Equal(t, int64(math.MaxInt64), results[0])

		// Test min value
		results, err = instance.Call("i64_identity", int64(math.MinInt64))
		require.NoError(t, err)
		assert.Equal(t, int64(math.MinInt64), results[0])

		// Test zero
		results, err = instance.Call("i64_identity", int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), results[0])
	})

	// ========================================
	// f32 Tests
	// ========================================
	t.Run("f32_identity", func(t *testing.T) {
		results, err := instance.Call("f32_identity", float32(3.14159))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(3.14159), results[0].(float32), 0.00001)
	})

	t.Run("f32_add", func(t *testing.T) {
		results, err := instance.Call("f32_add", float32(1.5), float32(2.5))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(4.0), results[0].(float32), 0.00001)
	})

	t.Run("f32_multiply", func(t *testing.T) {
		results, err := instance.Call("f32_multiply", float32(2.5), float32(4.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(10.0), results[0].(float32), 0.00001)
	})

	t.Run("f32_sqrt", func(t *testing.T) {
		results, err := instance.Call("f32_sqrt", float32(16.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float32(4.0), results[0].(float32), 0.00001)
	})

	t.Run("f32_edge_cases", func(t *testing.T) {
		// Test zero
		results, err := instance.Call("f32_identity", float32(0.0))
		require.NoError(t, err)
		assert.Equal(t, float32(0.0), results[0])

		// Test negative zero
		results, err = instance.Call("f32_identity", float32(math.Copysign(0, -1)))
		require.NoError(t, err)
		assert.Equal(t, float32(math.Copysign(0, -1)), results[0])

		// Test positive infinity
		results, err = instance.Call("f32_identity", float32(math.Inf(1)))
		require.NoError(t, err)
		assert.True(t, math.IsInf(float64(results[0].(float32)), 1))

		// Test negative infinity
		results, err = instance.Call("f32_identity", float32(math.Inf(-1)))
		require.NoError(t, err)
		assert.True(t, math.IsInf(float64(results[0].(float32)), -1))

		// Test NaN
		results, err = instance.Call("f32_identity", float32(math.NaN()))
		require.NoError(t, err)
		assert.True(t, math.IsNaN(float64(results[0].(float32))))

		// Test very small number
		results, err = instance.Call("f32_identity", float32(1.4e-45))
		require.NoError(t, err)
		assert.InDelta(t, float32(1.4e-45), results[0].(float32), 1e-50)

		// Test very large number
		results, err = instance.Call("f32_identity", float32(3.4e38))
		require.NoError(t, err)
		assert.InDelta(t, float32(3.4e38), results[0].(float32), 1e33)
	})

	// ========================================
	// f64 Tests
	// ========================================
	t.Run("f64_identity", func(t *testing.T) {
		results, err := instance.Call("f64_identity", float64(3.141592653589793))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(3.141592653589793), results[0])
	})

	t.Run("f64_add", func(t *testing.T) {
		results, err := instance.Call("f64_add", float64(1.5), float64(2.5))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(4.0), results[0])
	})

	t.Run("f64_multiply", func(t *testing.T) {
		results, err := instance.Call("f64_multiply", float64(2.5), float64(4.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(10.0), results[0])
	})

	t.Run("f64_sqrt", func(t *testing.T) {
		results, err := instance.Call("f64_sqrt", float64(16.0))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(4.0), results[0])
	})

	t.Run("f64_edge_cases", func(t *testing.T) {
		// Test zero
		results, err := instance.Call("f64_identity", float64(0.0))
		require.NoError(t, err)
		assert.Equal(t, float64(0.0), results[0])

		// Test negative zero
		results, err = instance.Call("f64_identity", math.Copysign(0, -1))
		require.NoError(t, err)
		assert.Equal(t, math.Copysign(0, -1), results[0])

		// Test positive infinity
		results, err = instance.Call("f64_identity", math.Inf(1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(results[0].(float64), 1))

		// Test negative infinity
		results, err = instance.Call("f64_identity", math.Inf(-1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(results[0].(float64), -1))

		// Test NaN
		results, err = instance.Call("f64_identity", math.NaN())
		require.NoError(t, err)
		assert.True(t, math.IsNaN(results[0].(float64)))

		// Test very small number
		results, err = instance.Call("f64_identity", float64(5e-324))
		require.NoError(t, err)
		assert.Equal(t, float64(5e-324), results[0])

		// Test very large number
		results, err = instance.Call("f64_identity", float64(1.7e308))
		require.NoError(t, err)
		assert.InDelta(t, float64(1.7e308), results[0].(float64), 1e303)
	})

	// ========================================
	// Mixed Type Tests
	// ========================================
	t.Run("mixed_params", func(t *testing.T) {
		// Test: (i32 + i64) * f32 + f64
		// (10 + 5) * 2.0 + 0.5 = 30.5
		results, err := instance.Call("mixed_params",
			int32(10),
			int64(5),
			float32(2.0),
			float64(0.5))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.InDelta(t, float64(30.5), results[0].(float64), 0.001)
	})

	t.Run("multi_return", func(t *testing.T) {
		results, err := instance.Call("multi_return")
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Check each return value type and value
		assert.Equal(t, int32(42), results[0])
		assert.InDelta(t, float64(3.14159), results[1].(float64), 0.00001)
		assert.Equal(t, int64(9223372036854775807), results[2])
	})

	t.Run("all_types_return", func(t *testing.T) {
		results, err := instance.Call("all_types_return")
		require.NoError(t, err)
		require.Len(t, results, 4)

		// Check all four types
		assert.Equal(t, int32(100), results[0])
		assert.Equal(t, int64(200), results[1])
		assert.InDelta(t, float32(1.5), results[2].(float32), 0.001)
		assert.InDelta(t, float64(2.5), results[3].(float64), 0.001)
	})

	// ========================================
	// Type Conversion Tests
	// ========================================
	t.Run("i32_to_f64", func(t *testing.T) {
		results, err := instance.Call("i32_to_f64", int32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, float64(42.0), results[0])
	})

	t.Run("f64_to_i32", func(t *testing.T) {
		results, err := instance.Call("f64_to_i32", float64(42.7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), results[0])
	})

	// ========================================
	// Precision Tests
	// ========================================
	t.Run("precision_preservation", func(t *testing.T) {
		// Verify that f64 maintains precision
		preciseValue := float64(1.2345678901234567)
		results, err := instance.Call("f64_identity", preciseValue)
		require.NoError(t, err)
		assert.Equal(t, preciseValue, results[0])

		// Verify f32 precision limits
		f32Value := float32(1.234567)
		results, err = instance.Call("f32_identity", f32Value)
		require.NoError(t, err)
		// f32 has ~7 decimal digits of precision
		assert.InDelta(t, f32Value, results[0].(float32), 1e-6)
	})
}
