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

	// Create a linker and define WASI
	linker, err := wasmtime.NewLinker(engine)
	if err != nil {
		log.Fatalf("Failed to create linker: %v", err)
	}
	defer linker.Close()

	if err := linker.DefineWASI(); err != nil {
		log.Fatalf("Failed to define WASI: %v", err)
	}
	fmt.Println("✓ Defined WASI in linker")

	// Instantiate the module using the linker (it will provide WASI imports)
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		log.Fatalf("Failed to instantiate module: %v", err)
	}
	fmt.Println("✓ Instantiated module")

	// Call the _start function
	// No wrapper needed! instance.Call now automatically treats WASI exit(0) as success
	fmt.Println("\nCalling _start function:")
	_, err = instance.Call("_start")
	if err != nil {
		log.Fatalf("Failed to call _start: %v", err)
	}

	fmt.Println("\n✓ WASI program executed successfully!")
}
