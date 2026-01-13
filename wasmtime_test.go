package wasmtime_test

import (
	"testing"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func TestEngineCreation(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()
}

func TestStoreCreation(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()
}

func TestWATCompilation(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
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
	if err != nil {
		t.Fatalf("failed to compile WAT: %v", err)
	}
	defer module.Close()
}

func TestSimpleFunctionCall(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/simple.wat")
	if err != nil {
		t.Fatalf("failed to load module: %v", err)
	}
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	if err != nil {
		t.Fatalf("failed to instantiate: %v", err)
	}

	// Test that we can get the export
	_, err = instance.GetExport("add")
	if err != nil {
		t.Fatalf("failed to get export: %v", err)
	}

	t.Log("Successfully instantiated module and accessed exports")
}

func TestWASIHello(t *testing.T) {
	engine, err := wasmtime.NewEngine()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	store, err := wasmtime.NewStore(engine)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Configure WASI
	wasiConfig, err := wasmtime.NewWASIConfig()
	if err != nil {
		t.Fatalf("failed to create WASI config: %v", err)
	}

	wasiConfig.WithInheritStdio()

	if err := wasiConfig.Apply(store); err != nil {
		t.Fatalf("failed to apply WASI config: %v", err)
	}

	module, err := wasmtime.NewModuleFromWATFile(engine, "testdata/hello_wasi.wat")
	if err != nil {
		t.Fatalf("failed to load module: %v", err)
	}
	defer module.Close()

	instance, err := wasmtime.NewInstance(store, module, nil)
	if err != nil {
		t.Fatalf("failed to instantiate: %v", err)
	}

	// Call the _start function
	_, err = instance.Call("_start")
	if err != nil {
		t.Fatalf("failed to call _start: %v", err)
	}

	t.Log("WASI hello world executed successfully")
}
