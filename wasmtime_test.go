package wasmtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func TestEngineCreation(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()
}

func TestStoreCreation(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()
}

func TestWATCompilation(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	wat := `
(module
  (func (export "add") (param i32 i32) (result i32)
    local.get 0
    local.get 1
    i32.add
  )
)
`

	module, err := wasmtime.NewModuleFromWAT(engine, wat)
	require.NoError(t, err)
	defer module.Close()
}

func TestSimpleFunctionCall(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/simple.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	// Test that we can get the export
	_, err = instance.GetExport("add")
	require.NoError(t, err)

	t.Log("Successfully instantiated module and accessed exports")
}

func TestAddFunction(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/simple.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	// Test add function: 5 + 7 = 12
	results, err := instance.Call("add", int32(5), int32(7))
	require.NoError(t, err)
	require.Len(t, results, 1)

	result, ok := results[0].(int32)
	require.True(t, ok, "result should be int32")
	assert.Equal(t, int32(12), result, "5 + 7 should equal 12")
}

func TestMultiplyFunction(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/simple.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	// Test multiply function: 6 * 7 = 42
	results, err := instance.Call("multiply", int32(6), int32(7))
	require.NoError(t, err)
	require.Len(t, results, 1)

	result, ok := results[0].(int32)
	require.True(t, ok, "result should be int32")
	assert.Equal(t, int32(42), result, "6 * 7 should equal 42")
}

func TestMultipleFunctionCalls(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/simple.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		function string
		args     []interface{}
		expected int32
	}{
		{"add positive", "add", []interface{}{int32(10), int32(20)}, 30},
		{"add negative", "add", []interface{}{int32(-5), int32(15)}, 10},
		{"add zeros", "add", []interface{}{int32(0), int32(0)}, 0},
		{"multiply positive", "multiply", []interface{}{int32(3), int32(4)}, 12},
		{"multiply by zero", "multiply", []interface{}{int32(5), int32(0)}, 0},
		{"multiply negative", "multiply", []interface{}{int32(-2), int32(3)}, -6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := instance.Call(tc.function, tc.args...)
			require.NoError(t, err)
			require.Len(t, results, 1)

			result, ok := results[0].(int32)
			require.True(t, ok, "result should be int32")
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWASIHello(t *testing.T) {
	t.Skip("WASI test requires import linking - bindings ready but not yet implemented")

	engine, err := wasmtime.NewEngine()
	require.NoError(t, err)
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	require.NoError(t, err)
	defer store.Close()

	// Configure WASI
	wasiConfig, err := wasmtime.NewWASIConfig()
	require.NoError(t, err)

	wasiConfig.WithInheritStdio()

	err = wasiConfig.Apply(store)
	require.NoError(t, err)

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/hello_wasi.wat")
	require.NoError(t, err)
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)

	// Call the _start function
	_, err = instance.Call("_start")
	require.NoError(t, err)

	t.Log("WASI hello world executed successfully")
}
