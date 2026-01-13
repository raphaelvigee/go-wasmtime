package wasmtime_test

import (
	"testing"

	wasmtime "github.com/rvigee/purego-wasmtime"
	"github.com/stretchr/testify/require"
)

// TestMultiValueReturns tests functions returning 0, 1, and 2+ values
func TestMultiValueReturns(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

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

	module, err := wasmtime.NewModuleFromWAT(engine, wat)
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	t.Run("zero results", func(t *testing.T) {
		results, err := instance.Call("no_return", int32(5))
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d: %v", len(results), results)
		}
		t.Logf("✓ Void function returned %d results", len(results))
	})

	t.Run("one result", func(t *testing.T) {
		results, err := instance.Call("one_return", int32(5))
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0].(int32) != 15 {
			t.Errorf("Expected 15, got %v", results[0])
		}
		t.Logf("✓ Single-value function returned %d results: %v", len(results), results)
	})

	t.Run("two results", func(t *testing.T) {
		results, err := instance.Call("two_returns", int32(3), int32(4))
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		sum := results[0].(int32)
		product := results[1].(int32)

		if sum != 7 {
			t.Errorf("Expected sum=7, got %d", sum)
		}
		if product != 12 {
			t.Errorf("Expected product=12, got %d", product)
		}
		t.Logf("✓ Multi-value function returned %d results: sum=%d, product=%d", len(results), sum, product)
	})
}
