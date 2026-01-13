package main

import (
	"context"
	"fmt"
	"log"
	"os"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
	fmt.Println("=== WASI Example ===")

	ctx := context.Background()

	// Create runtime with WASI configuration
	config := wasmtime.NewRuntimeConfig().
		WithWASI(
			wasmtime.NewWASIConfig().
				WithInheritStdio().
				WithEnv("GREETING", "Hello from WASI!").
				WithEnv("USER", "wasmtime-go").
				WithArgs("wasi-program", "--hello"),
		)

	r, err := wasmtime.NewRuntimeWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer r.Close(ctx)
	fmt.Println("✓ Created runtime with WASI config")

	// Load and compile the WASI module
	wat, err := os.ReadFile("../testdata/hello_wasi.wat")
	if err != nil {
		log.Fatalf("Failed to read WAT file: %v", err)
	}

	compiled, err := r.CompileModule(ctx, wat)
	if err != nil {
		log.Fatalf("Failed to compile module: %v", err)
	}
	defer compiled.Close()
	fmt.Println("✓ Compiled WASI module")

	// Instantiate with WASI support
	mod, err := r.InstantiateWithWASI(ctx, compiled)
	if err != nil {
		log.Fatalf("Failed to instantiate module: %v", err)
	}
	defer mod.Close(ctx)
	fmt.Println("✓ Instantiated module with WASI")

	// Call the _start function
	fmt.Println("\nCalling _start function:")
	startFn := mod.ExportedFunction("_start")
	if startFn == nil {
		log.Fatal("_start function not found")
	}

	_, err = startFn.Call(ctx)
	if err != nil {
		log.Fatalf("Failed to call _start: %v", err)
	}

	fmt.Println("\n✓ WASI program executed successfully!")
}
