package wasmtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

// TestParameterCounts tests functions with 0, 1, and 2 parameters
func TestParameterCounts(t *testing.T) {
	ts := newTestSetup(t, "testdata/params.wat", false)
	defer ts.cleanup()

	t.Run("zero_parameters", func(t *testing.T) {
		fn := ts.module.ExportedFunction("constant")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(100), wasmtime.DecodeI32(results[0]))
	})

	t.Run("one_parameter_negate", func(t *testing.T) {
		fn := ts.module.ExportedFunction("negate")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(-42), wasmtime.DecodeI32(results[0]))
	})

	t.Run("one_parameter_double", func(t *testing.T) {
		fn := ts.module.ExportedFunction("double")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(21))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), wasmtime.DecodeI32(results[0]))
	})

	t.Run("one_parameter_square", func(t *testing.T) {
		fn := ts.module.ExportedFunction("square")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(49), wasmtime.DecodeI32(results[0]))
	})

	t.Run("two_parameters_subtract", func(t *testing.T) {
		fn := ts.module.ExportedFunction("subtract")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(100), wasmtime.EncodeI32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(58), wasmtime.DecodeI32(results[0]))
	})

	t.Run("two_parameters_max", func(t *testing.T) {
		fn := ts.module.ExportedFunction("max")
		require.NotNil(t, fn)

		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(10), wasmtime.EncodeI32(20))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(20), wasmtime.DecodeI32(results[0]))

		results, err = fn.Call(ts.ctx, wasmtime.EncodeI32(30), wasmtime.EncodeI32(15))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(30), wasmtime.DecodeI32(results[0]))
	})

	t.Run("edge_cases", func(t *testing.T) {
		// Test with zero
		fn := ts.module.ExportedFunction("negate")
		results, err := fn.Call(ts.ctx, wasmtime.EncodeI32(0))
		require.NoError(t, err)
		assert.Equal(t, int32(0), wasmtime.DecodeI32(results[0]))

		// Test with negative numbers
		fn = ts.module.ExportedFunction("double")
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI32(-10))
		require.NoError(t, err)
		assert.Equal(t, int32(-20), wasmtime.DecodeI32(results[0]))

		// Test subtraction with negatives
		fn = ts.module.ExportedFunction("subtract")
		results, err = fn.Call(ts.ctx, wasmtime.EncodeI32(-5), wasmtime.EncodeI32(-10))
		require.NoError(t, err)
		assert.Equal(t, int32(5), wasmtime.DecodeI32(results[0]))
	})
}
