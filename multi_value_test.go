package wasmtime_test

import (
	"context"
	"testing"

	wasmtime "github.com/rvigee/purego-wasmtime"
	"github.com/stretchr/testify/require"
)

// TestMultiValueReturns tests functions returning 0, 1, and 2+ values
func TestMultiValueReturns(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)

	// WAT with various return counts
	wat := `
		(module
			;; 0 results - void function
			(func $no_return (param $a i32)
				;; Just do nothing, no return
			)
			(export "no_return" (func $no_return))

			;; 1 result - standard function
			(func $one_return (param $a i32) (result i32)
				local.get $a
				i32.const 10
				i32.add
			)
			(export "one_return" (func $one_return))

			;; 2 results - multi-value return (swap and add)
			(func $two_returns (param $a i32) (param $b i32) (result i32 i32)
				;; First result: sum
				local.get $a
				local.get $b
				i32.add
				;; Second result: product
				local.get $a
				local.get $b
				i32.mul
			)
			(export "two_returns" (func $two_returns))
		)
	`

	// Compile inline WAT
	compiled, err := r.CompileModule(ctx, []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(ctx, compiled)
	require.NoError(t, err)
	defer mod.Close(ctx)

	t.Run("zero results", func(t *testing.T) {
		fn := mod.ExportedFunction("no_return")
		require.NotNil(t, fn)

		results, err := fn.Call(ctx, wasmtime.EncodeI32(5))
		require.NoError(t, err)
		require.Len(t, results, 0, "Expected 0 results")
	})

	t.Run("one result", func(t *testing.T) {
		fn := mod.ExportedFunction("one_return")
		require.NotNil(t, fn)

		results, err := fn.Call(ctx, wasmtime.EncodeI32(5))
		require.NoError(t, err)
		require.Len(t, results, 1, "Expected 1 result")
		require.Equal(t, int32(15), wasmtime.DecodeI32(results[0]))
	})

	t.Run("two results", func(t *testing.T) {
		fn := mod.ExportedFunction("two_returns")
		require.NotNil(t, fn)

		results, err := fn.Call(ctx, wasmtime.EncodeI32(3), wasmtime.EncodeI32(4))
		require.NoError(t, err)
		require.Len(t, results, 2, "Expected 2 results")

		sum := wasmtime.DecodeI32(results[0])
		product := wasmtime.DecodeI32(results[1])

		require.Equal(t, int32(7), sum, "Expected sum=7")
		require.Equal(t, int32(12), product, "Expected product=12")
	})
}
