package benchmark

import (
	"fmt"
	"os"
	"testing"

	_ "embed"

	"github.com/bytecodealliance/wasmtime-go/v28"
	purego "github.com/rvigee/purego-wasmtime"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/wasmerio/wasmer-go/wasmer"
)

var (
	//go:embed fib.wat
	fibWat []byte

	fibWasm []byte

	fibIterations = 25
)

func TestMain(m *testing.M) {
	var err error
	fibWasm, err = wasmtime.Wat2Wasm(string(fibWat))
	if err != nil {
		fmt.Printf("failed to compile wat to wasm: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// --- Startup Benchmarks (NewRuntime + Compile + Instantiate) ---

func BenchmarkWazero_Startup(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		r := wazero.NewRuntime(ctx)
		compiled, err := r.CompileModule(ctx, fibWasm)
		require.NoError(b, err)

		mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
		if err != nil {
			r.Close(ctx)
			require.NoError(b, err)
		}
		mod.Close(ctx)
		r.Close(ctx)
	}
}

func BenchmarkPurego_Startup(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		runtime, err := purego.NewRuntime(ctx)
		require.NoError(b, err)

		mod, err := runtime.CompileModule(ctx, fibWasm)
		if err != nil {
			runtime.Close(ctx)
			require.NoError(b, err)
		}

		instance, err := runtime.Instantiate(ctx, mod)
		if err != nil {
			mod.Close()
			runtime.Close(ctx)
			require.NoError(b, err)
		}

		mod.Close()
		runtime.Close(ctx)
		_ = instance
	}
}

func BenchmarkWasmtimeGo_Startup(b *testing.B) {
	for b.Loop() {
		engine := wasmtime.NewEngine()
		mod, err := wasmtime.NewModule(engine, fibWasm)
		require.NoError(b, err)

		store := wasmtime.NewStore(engine)
		_, err = wasmtime.NewInstance(store, mod, []wasmtime.AsExtern{})
		require.NoError(b, err)

		// Finalizers handle cleanup
		_ = store
		_ = engine
	}
}

// --- Wazero ---

func BenchmarkWazero_Compile(b *testing.B) {
	ctx := b.Context()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	for b.Loop() {
		_, err := r.CompileModule(ctx, fibWasm)
		require.NoError(b, err)
	}
}

func BenchmarkWazero_Instantiate(b *testing.B) {
	ctx := b.Context()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)
	compiled, err := r.CompileModule(ctx, fibWasm)
	require.NoError(b, err)

	// wazero compiled module doesn't need close

	for b.Loop() {
		mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
		require.NoError(b, err)
		mod.Close(ctx)
	}
}

func BenchmarkWazero_Exec(b *testing.B) {
	ctx := b.Context()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)
	compiled, err := r.CompileModule(ctx, fibWasm)
	require.NoError(b, err)

	mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	require.NoError(b, err)

	fib := mod.ExportedFunction("fib")

	b.ResetTimer()

	for b.Loop() {
		res, err := fib.Call(ctx, uint64(fibIterations))
		require.NoError(b, err)
		require.Equal(b, uint64(75025), res[0])
	}
}

// --- Wasmtime Purego ---

func BenchmarkPurego_Compile(b *testing.B) {
	ctx := b.Context()
	runtime, err := purego.NewRuntime(ctx)
	require.NoError(b, err)
	defer runtime.Close(ctx)

	for b.Loop() {
		mod, err := runtime.CompileModule(ctx, fibWasm)
		require.NoError(b, err)
		mod.Close()
	}
}

// BenchmarkPurego_Instantiate skipped due to instance cleanup limitations in current API

func BenchmarkPurego_Exec(b *testing.B) {
	ctx := b.Context()
	runtime, err := purego.NewRuntime(ctx)
	require.NoError(b, err)
	defer runtime.Close(ctx)

	mod, err := runtime.CompileModule(ctx, fibWasm)
	require.NoError(b, err)
	defer mod.Close()

	instance, err := runtime.Instantiate(ctx, mod)
	require.NoError(b, err)

	fib := instance.ExportedFunction("fib")
	require.NotNil(b, fib)

	b.ResetTimer()

	for b.Loop() {
		res, err := fib.Call(ctx, uint64(fibIterations))
		require.NoError(b, err)
		require.Equal(b, uint64(75025), res[0])
	}
}

// --- Wasmtime Go (Cgo) ---

func BenchmarkWasmtimeGo_Compile(b *testing.B) {
	engine := wasmtime.NewEngine()

	for b.Loop() {
		_, err := wasmtime.NewModule(engine, fibWasm)
		require.NoError(b, err)
	}
}

func BenchmarkWasmtimeGo_Instantiate(b *testing.B) {
	engine := wasmtime.NewEngine()
	mod, err := wasmtime.NewModule(engine, fibWasm)
	require.NoError(b, err)

	for b.Loop() {
		store := wasmtime.NewStore(engine)
		_, err := wasmtime.NewInstance(store, mod, []wasmtime.AsExtern{})
		require.NoError(b, err)
		// Usually internal store is enough for cleanup
	}
}

func BenchmarkWasmtimeGo_Exec(b *testing.B) {
	engine := wasmtime.NewEngine()
	mod, err := wasmtime.NewModule(engine, fibWasm)
	require.NoError(b, err)

	store := wasmtime.NewStore(engine)
	instance, err := wasmtime.NewInstance(store, mod, []wasmtime.AsExtern{})
	require.NoError(b, err)

	fib := instance.GetFunc(store, "fib")

	b.ResetTimer()

	for b.Loop() {
		res, err := fib.Call(store, fibIterations)
		require.NoError(b, err)
		require.Equal(b, int32(75025), res.(int32))
	}
}

// --- Wasmer ---

func BenchmarkWasmer_Startup(b *testing.B) {
	for b.Loop() {
		engine := wasmer.NewEngine()
		store := wasmer.NewStore(engine)
		mod, err := wasmer.NewModule(store, fibWasm)
		require.NoError(b, err)

		instance, err := wasmer.NewInstance(mod, wasmer.NewImportObject())
		require.NoError(b, err)
		instance.Close()
	}
}

func BenchmarkWasmer_Compile(b *testing.B) {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	for b.Loop() {
		_, err := wasmer.NewModule(store, fibWasm)
		require.NoError(b, err)
	}
}

func BenchmarkWasmer_Instantiate(b *testing.B) {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)
	mod, err := wasmer.NewModule(store, fibWasm)
	require.NoError(b, err)
	importObject := wasmer.NewImportObject()

	for b.Loop() {
		instance, err := wasmer.NewInstance(mod, importObject)
		require.NoError(b, err)
		instance.Close()
	}
}

func BenchmarkWasmer_Exec(b *testing.B) {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)
	mod, err := wasmer.NewModule(store, fibWasm)
	require.NoError(b, err)

	instance, err := wasmer.NewInstance(mod, wasmer.NewImportObject())
	require.NoError(b, err)

	fib, err := instance.Exports.GetFunction("fib")
	require.NoError(b, err)

	b.ResetTimer()

	for b.Loop() {
		res, err := fib(fibIterations)
		require.NoError(b, err)
		require.Equal(b, int32(75025), res)
	}
}
