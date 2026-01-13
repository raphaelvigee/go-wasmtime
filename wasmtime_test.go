package wasmtime_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func TestRuntimeCreation(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)
}

func TestWATCompilation(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)

	wat := `
(module
  (func (export "add") (param i32 i32) (result i32)
    local.get 0
    local.get 1
    i32.add
  )
)
`

	compiled, err := r.CompileModule(ctx, []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()
}

func TestSimpleFunctionCall(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)

	// Read WAT file
	const watPath = "testdata/simple.wat"
	wat, err := readFile(watPath)
	require.NoError(t, err)

	compiled, err := r.CompileModule(ctx, wat)
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(ctx, compiled)
	require.NoError(t, err)
	defer mod.Close(ctx)

	// Test that we can get the export
	fn := mod.ExportedFunction("add")
	require.NotNil(t, fn)

	t.Log("Successfully instantiated module and accessed exports")
}

func TestAddFunction(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)

	wat, err := readFile("testdata/simple.wat")
	require.NoError(t, err)

	compiled, err := r.CompileModule(ctx, wat)
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(ctx, compiled)
	require.NoError(t, err)
	defer mod.Close(ctx)

	// Test add function: 5 + 7 = 12
	addFn := mod.ExportedFunction("add")
	require.NotNil(t, addFn)

	results, err := addFn.Call(ctx, wasmtime.EncodeI32(5), wasmtime.EncodeI32(7))
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := wasmtime.DecodeI32(results[0])
	assert.Equal(t, int32(12), result, "5 + 7 should equal 12")
}

func TestMultiplyFunction(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)

	wat, err := readFile("testdata/simple.wat")
	require.NoError(t, err)

	compiled, err := r.CompileModule(ctx, wat)
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(ctx, compiled)
	require.NoError(t, err)
	defer mod.Close(ctx)

	// Test multiply function: 6 * 7 = 42
	mulFn := mod.ExportedFunction("multiply")
	require.NotNil(t, mulFn)

	results, err := mulFn.Call(ctx, wasmtime.EncodeI32(6), wasmtime.EncodeI32(7))
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := wasmtime.DecodeI32(results[0])
	assert.Equal(t, int32(42), result, "6 * 7 should equal 42")
}

func TestMultipleFunctionCalls(t *testing.T) {
	ctx := context.Background()
	r, err := wasmtime.NewRuntime(ctx)
	require.NoError(t, err)
	defer r.Close(ctx)

	wat, err := readFile("testdata/simple.wat")
	require.NoError(t, err)

	compiled, err := r.CompileModule(ctx, wat)
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(ctx, compiled)
	require.NoError(t, err)
	defer mod.Close(ctx)

	testCases := []struct {
		name     string
		function string
		arg1     int32
		arg2     int32
		expected int32
	}{
		{"add positive", "add", 10, 20, 30},
		{"add negative", "add", -5, 15, 10},
		{"add zeros", "add", 0, 0, 0},
		{"multiply positive", "multiply", 3, 4, 12},
		{"multiply by zero", "multiply", 5, 0, 0},
		{"multiply negative", "multiply", -2, 3, -6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := mod.ExportedFunction(tc.function)
			require.NotNil(t, fn)

			results, err := fn.Call(ctx, wasmtime.EncodeI32(tc.arg1), wasmtime.EncodeI32(tc.arg2))
			require.NoError(t, err)
			require.Len(t, results, 1)

			result := wasmtime.DecodeI32(results[0])
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWASIHello(t *testing.T) {
	ctx := context.Background()

	// Create runtime with WASI configuration
	config := wasmtime.NewRuntimeConfig().
		WithWASI(
			wasmtime.NewWASIConfig().
				WithInheritStdio(),
		)

	r, err := wasmtime.NewRuntimeWithConfig(ctx, config)
	require.NoError(t, err)
	defer r.Close(ctx)

	wat, err := readFile("testdata/hello_wasi.wat")
	require.NoError(t, err)

	compiled, err := r.CompileModule(ctx, wat)
	require.NoError(t, err)
	defer compiled.Close()

	// Use InstantiateWithWASI for WASI modules
	mod, err := r.InstantiateWithWASI(ctx, compiled)
	require.NoError(t, err)
	defer mod.Close(ctx)

	// Call the _start function
	startFn := mod.ExportedFunction("_start")
	require.NotNil(t, startFn)

	_, err = startFn.Call(ctx)
	require.NoError(t, err)

	t.Log("WASI hello world executed successfully")
}

// Helper function to read file
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
