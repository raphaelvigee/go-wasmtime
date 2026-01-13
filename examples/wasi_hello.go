package main

import (
	"fmt"
	"log"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
	fmt.Println("=== WASI Example ===")

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

	// Configure WASI
	wasiConfig, err := wasmtime.NewWASIConfig()
	if err != nil {
		log.Fatalf("Failed to create WASI config: %v", err)
	}

	// Inherit stdio from host
	wasiConfig.WithInheritStdio()

	// Set environment variables
	wasiConfig.WithEnv(map[string]string{
		"GREETING": "Hello from WASI!",
		"USER":     "wasmtime-go",
	})

	// Set command-line arguments
	wasiConfig.WithArgs([]string{"wasi-program", "--hello"})

	// Apply WASI configuration to store
	if err := wasiConfig.Apply(store); err != nil {
		log.Fatalf("Failed to apply WASI config: %v", err)
	}
	fmt.Println("✓ Configured WASI with stdio, env, and args")

	// Load and compile the WASI module
	module, err := wasmtime.NewModuleFromWATFile(engine, "../testdata/hello_wasi.wat")
	if err != nil {
		log.Fatalf("Failed to load module: %v", err)
	}
	defer module.Close()
	fmt.Println("✓ Compiled WASI module")

	// Instantiate the module
	instance, err := wasmtime.NewInstance(store, module, nil)
	if err != nil {
		log.Fatalf("Failed to instantiate module: %v", err)
	}
	fmt.Println("✓ Instantiated module")

	// Call the _start function
	fmt.Println("\nCalling _start function:")
	_, err = instance.Call("_start")
	if err != nil {
		log.Fatalf("Failed to call _start: %v", err)
	}

	fmt.Println("\n✓ WASI program executed successfully!")
}
