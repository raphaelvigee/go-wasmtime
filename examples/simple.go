package main

import (
	"fmt"
	"log"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
	fmt.Println("=== Simple Wasmtime Example ===")

	// Create an engine
	engine, err := wasmtime.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	fmt.Println("✓ Created engine")

	// Create a store
	store, err := wasmtime.NewStore(engine)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	fmt.Println("✓ Created store")

	// Define a simple WAT module
	wat := `
(module
  (func (export "add") (param i32 i32) (result i32)
    local.get 0
    local.get 1
    i32.add
  )
  (func (export "subtract") (param i32 i32) (result i32)
    local.get 0
    local.get 1
    i32.sub
  )
)
`

	// Compile the module
	module, err := wasmtime.NewModuleFromWAT(engine, wat)
	if err != nil {
		log.Fatalf("Failed to compile module: %v", err)
	}
	defer module.Close()
	fmt.Println("✓ Compiled WAT module")

	// Instantiate the module
	instance, err := wasmtime.NewInstance(store, module, nil)
	if err != nil {
		log.Fatalf("Failed to instantiate module: %v", err)
	}
	fmt.Println("✓ Instantiated module")

	// Get exported functions
	addFunc, err := instance.GetExport("add")
	if err != nil {
		log.Fatalf("Failed to get 'add' export: %v", err)
	}
	fmt.Println("✓ Found 'add' function export")

	subFunc, err := instance.GetExport("subtract")
	if err != nil {
		log.Fatalf("Failed to get 'subtract' export: %v", err)
	}
	fmt.Println("✓ Found 'subtract' function export")

	_ = addFunc
	_ = subFunc

	fmt.Println("\nSuccessfully created and instantiated WebAssembly module!")
	fmt.Println("Functions 'add' and 'subtract' are ready to use.")
}
