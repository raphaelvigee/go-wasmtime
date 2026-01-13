package wasmtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func TestZeroArgsFunction(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/zero_args.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	// Test calling a function with NO arguments
	t.Log("Calling get_forty_two() with no arguments...")
	results, err := instance.Call("get_forty_two")
	if err != nil {
		t.Fatalf("Failed to call function: %v", err)
	}
	t.Logf("Results: %v", results)
}
