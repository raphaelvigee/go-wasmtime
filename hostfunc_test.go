package wasmtime

import (
	"context"
	"testing"

	"github.com/rvigee/purego-wasmtime/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock GoFunction implementation for testing
type addFunc struct{}

func (a *addFunc) Call(ctx context.Context, params []uint64) ([]uint64, error) {
	if len(params) != 2 {
		return nil, nil
	}
	sum := DecodeI32(params[0]) + DecodeI32(params[1])
	return []uint64{EncodeI32(sum)}, nil
}

func TestHostFunctionBuilder(t *testing.T) {
	require.NoError(t, Initialize())

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	t.Run("create_host_module_builder", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("env")
		require.NotNil(t, builder)

		hmb, ok := builder.(*hostModuleBuilder)
		require.True(t, ok, "Builder is not *hostModuleBuilder")
		assert.Equal(t, "env", hmb.moduleName)
	})

	t.Run("add_function_with_goFunc", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("env")

		called := false
		hostFunc := func(ctx context.Context, stack []uint64) {
			called = true
			if len(stack) >= 2 {
				stack[1] = stack[0]
			}
		}

		builder.NewFunctionBuilder("echo",
			[]api.ValueType{api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).WithGoFunc(hostFunc).Export("echo")

		hmb := builder.(*hostModuleBuilder)
		require.Len(t, hmb.functions, 1)

		fn := hmb.functions[0]
		assert.Equal(t, "echo", fn.name)
		assert.NotNil(t, fn.goFunc)
		assert.Len(t, fn.paramTypes, 1)
		assert.Equal(t, api.ValueTypeI32, fn.paramTypes[0])
		assert.Len(t, fn.resultTypes, 1)
		assert.Equal(t, api.ValueTypeI32, fn.resultTypes[0])

		_ = called
	})

	t.Run("add_function_with_goFunction", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("env")

		addFn := &addFunc{}

		builder.NewFunctionBuilder("add",
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).WithGoFunction(addFn).
			WithParameterNames("a", "b").
			WithResultNames("sum").
			Export("add")

		hmb := builder.(*hostModuleBuilder)
		fn := hmb.functions[0]

		assert.Len(t, fn.paramNames, 2)
		assert.Equal(t, []string{"a", "b"}, fn.paramNames)
		assert.Len(t, fn.resultNames, 1)
		assert.Equal(t, []string{"sum"}, fn.resultNames)
	})

	t.Run("multiple_functions", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("math")

		builder.NewFunctionBuilder("add",
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			a := DecodeI32(stack[0])
			b := DecodeI32(stack[1])
			stack[2] = EncodeI32(a + b)
		}).Export("add")

		builder.NewFunctionBuilder("multiply",
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			a := DecodeI32(stack[0])
			b := DecodeI32(stack[1])
			stack[2] = EncodeI32(a * b)
		}).Export("multiply")

		builder.NewFunctionBuilder("negate",
			[]api.ValueType{api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			val := DecodeI32(stack[0])
			stack[1] = EncodeI32(-val)
		}).Export("negate")

		hmb := builder.(*hostModuleBuilder)
		require.Len(t, hmb.functions, 3)

		names := []string{}
		for _, fn := range hmb.functions {
			names = append(names, fn.name)
		}
		assert.Equal(t, []string{"add", "multiply", "negate"}, names)
	})

	t.Run("function_with_no_params", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("env")

		builder.NewFunctionBuilder("get_constant",
			[]api.ValueType{},
			[]api.ValueType{api.ValueTypeI32},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			stack[0] = EncodeI32(42)
		}).Export("get_constant")

		hmb := builder.(*hostModuleBuilder)
		fn := hmb.functions[0]

		assert.Empty(t, fn.paramTypes)
		assert.Len(t, fn.resultTypes, 1)
	})

	t.Run("function_with_no_results", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("env")

		builder.NewFunctionBuilder("log",
			[]api.ValueType{api.ValueTypeI32},
			[]api.ValueType{},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			_ = DecodeI32(stack[0])
		}).Export("log")

		hmb := builder.(*hostModuleBuilder)
		fn := hmb.functions[0]

		assert.Empty(t, fn.resultTypes)
	})

	t.Run("instantiate_host_module", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("env")

		builder.NewFunctionBuilder("noop",
			[]api.ValueType{},
			[]api.ValueType{},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			// No-op
		}).Export("noop")

		assert.NoError(t, builder.Instantiate(t.Context()))
		builder.Close(t.Context())
	})

	t.Run("float_types", func(t *testing.T) {
		builder := r.NewHostModuleBuilder("math")

		builder.NewFunctionBuilder("add_f32",
			[]api.ValueType{api.ValueTypeF32, api.ValueTypeF32},
			[]api.ValueType{api.ValueTypeF32},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			a := DecodeF32(stack[0])
			b := DecodeF32(stack[1])
			stack[2] = EncodeF32(a + b)
		}).Export("add_f32")

		builder.NewFunctionBuilder("add_f64",
			[]api.ValueType{api.ValueTypeF64, api.ValueTypeF64},
			[]api.ValueType{api.ValueTypeF64},
		).WithGoFunc(func(ctx context.Context, stack []uint64) {
			a := DecodeF64(stack[0])
			b := DecodeF64(stack[1])
			stack[2] = EncodeF64(a + b)
		}).Export("add_f64")

		hmb := builder.(*hostModuleBuilder)
		assert.Len(t, hmb.functions, 2)
	})
}

func TestGoFunctionAdapter(t *testing.T) {
	addFn := &addFunc{}
	results, err := addFn.Call(t.Context(), []uint64{EncodeI32(5), EncodeI32(7)})
	require.NoError(t, err)
	require.Len(t, results, 1)

	sum := DecodeI32(results[0])
	assert.Equal(t, int32(12), sum)
}

func TestHostModuleBuilderChaining(t *testing.T) {
	require.NoError(t, Initialize())

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	builder := r.NewHostModuleBuilder("test")

	noop := func(ctx context.Context, stack []uint64) {}

	funcBuilder := builder.NewFunctionBuilder("test_func",
		[]api.ValueType{api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).WithGoFunc(noop).
		WithParameterNames("input").
		WithResultNames("output")

	require.NotNil(t, funcBuilder, "Function builder chain returned nil")

	funcBuilder.Export("test_func")

	hmb := builder.(*hostModuleBuilder)
	assert.Len(t, hmb.functions, 1, "Function not added after Export()")
}

func TestHostFunctionIntegration(t *testing.T) {

	require.NoError(t, Initialize())

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	// Define host functions that WASM will call
	hostModule := r.NewHostModuleBuilder("env")

	// Track calls to verify they were invoked
	addCalled := false
	logCalled := false
	var loggedValue int32

	// Host function: add two numbers
	hostModule.NewFunctionBuilder("host_add",
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).WithGoFunc(func(ctx context.Context, stack []uint64) {
		addCalled = true
		a := DecodeI32(stack[0])
		b := DecodeI32(stack[1])
		result := a + b
		stack[2] = EncodeI32(result)
		t.Logf("host_add(%d, %d) = %d", a, b, result)
	}).Export("host_add")

	// Host function: log a number (no return value)
	hostModule.NewFunctionBuilder("host_log",
		[]api.ValueType{api.ValueTypeI32},
		[]api.ValueType{},
	).WithGoFunc(func(ctx context.Context, stack []uint64) {
		logCalled = true
		loggedValue = DecodeI32(stack[0])
		t.Logf("host_log(%d) called", loggedValue)
	}).Export("host_log")

	// Instantiate the host module to make functions available
	require.NoError(t, hostModule.Instantiate(t.Context()))
	defer hostModule.Close(t.Context())

	// WASM module that imports and calls our host functions
	wat := `(module
		(import "env" "host_add" (func $host_add (param i32 i32) (result i32)))
		(import "env" "host_log" (func $host_log (param i32)))
		(func (export "test_host_functions") (result i32)
			(local i32)
			(i32.const 10)
			(i32.const 20)
			(call $host_add)
			(local.set 0)
			(local.get 0)
			(call $host_log)
			(local.get 0)
		)
	)`

	compiled, err := r.CompileModule(t.Context(), []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	// Instantiate the WASM module (should link with host functions)
	mod, err := r.Instantiate(t.Context(), compiled)
	require.NoError(t, err)
	defer mod.Close(t.Context())

	// Call the exported function
	fn := mod.ExportedFunction("test_host_functions")
	require.NotNil(t, fn)

	results, err := fn.Call(t.Context())
	require.NoError(t, err)
	require.Len(t, results, 1)

	// Verify the result (10 + 20 = 30)
	result := DecodeI32(results[0])
	assert.Equal(t, int32(30), result)

	// Verify host functions were actually called
	assert.True(t, addCalled, "host_add was not called")
	assert.True(t, logCalled, "host_log was not called")
	assert.Equal(t, int32(30), loggedValue, "Logged value incorrect")
}

func TestHostFunctionWithMemoryAccess(t *testing.T) {
	require.NoError(t, Initialize())

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	// Host function that reads from WASM memory
	hostModule := r.NewHostModuleBuilder("env")

	hostModule.NewFunctionBuilder("read_string",
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).WithGoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
		ptr := DecodeI32(stack[0])
		length := DecodeI32(stack[1])

		// Access the module's memory
		mem := mod.ExportedMemory("memory")
		require.NotNil(t, mem)

		data := mem.Data(ctx)

		// Read string from memory
		str := string((*[1 << 30]byte)(data)[ptr : ptr+length])
		t.Logf("Read string from WASM memory: %q", str)

		// Return the length
		stack[2] = EncodeI32(length)
	}).Export("read_string")

	require.NoError(t, hostModule.Instantiate(t.Context()))
	defer hostModule.Close(t.Context())

	// WASM module with memory and function that calls host
	wat := `
	(module
		(import "env" "read_string" (func $read_string (param i32 i32) (result i32)))
		
		(memory (export "memory") 1)
		(data (i32.const 0) "Hello from WASM!")
		
		(func (export "test") (result i32)
			(i32.const 0)   ;; pointer to string
			(i32.const 16)  ;; length
			(call $read_string)
		)
	)`

	compiled, err := r.CompileModule(t.Context(), []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(t.Context(), compiled)
	require.NoError(t, err)
	defer mod.Close(t.Context())

	fn := mod.ExportedFunction("test")
	require.NotNil(t, fn)

	results, err := fn.Call(t.Context())
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, int32(16), DecodeI32(results[0]))
}
