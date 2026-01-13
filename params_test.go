package wasmtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

// TestParameterCounts tests functions with 0, 1, and 2 parameters
func TestParameterCounts(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/params.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	t.Run("zero_parameters", func(t *testing.T) {
		results, err := instance.Call("constant")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(100), results[0])
	})

	t.Run("one_parameter_negate", func(t *testing.T) {
		results, err := instance.Call("negate", int32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(-42), results[0])
	})

	t.Run("one_parameter_double", func(t *testing.T) {
		results, err := instance.Call("double", int32(21))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(42), results[0])
	})

	t.Run("one_parameter_square", func(t *testing.T) {
		results, err := instance.Call("square", int32(7))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(49), results[0])
	})

	t.Run("two_parameters_subtract", func(t *testing.T) {
		results, err := instance.Call("subtract", int32(100), int32(42))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(58), results[0])
	})

	t.Run("two_parameters_max", func(t *testing.T) {
		results, err := instance.Call("max", int32(10), int32(20))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(20), results[0])

		results, err = instance.Call("max", int32(30), int32(15))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, int32(30), results[0])
	})

	t.Run("edge_cases", func(t *testing.T) {
		// Test with zero
		results, err := instance.Call("negate", int32(0))
		require.NoError(t, err)
		assert.Equal(t, int32(0), results[0])

		// Test with negative numbers
		results, err = instance.Call("double", int32(-10))
		require.NoError(t, err)
		assert.Equal(t, int32(-20), results[0])

		// Test subtraction with negatives
		results, err = instance.Call("subtract", int32(-5), int32(-10))
		require.NoError(t, err)
		assert.Equal(t, int32(5), results[0])
	})
}
