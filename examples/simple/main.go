package main

import (
	"context"
	"fmt"
	"log"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
	ctx := context.Background()

	// Create runtime
	r, err := wasmtime.NewRuntime(ctx)
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer r.Close(ctx)

	// Compile WAT module
	wat := `
	(module
	  (func (export "add") (param i32 i32) (result i32)
	    local.get 0
	    local.get 1
	    i32.add
	  )
	  (func (export "multiply") (param i32 i32) (result i32)
	    local.get 0
	    local.get 1
	    i32.mul
	  )
	)
	`

	compiled, err := r.CompileModule(ctx, []byte(wat))
	if err != nil {
		log.Fatalf("Failed to compile module: %v", err)
	}
	defer compiled.Close()

	// Instantiate module
	mod, err := r.Instantiate(ctx, compiled)
	if err != nil {
		log.Fatalf("Failed to instantiate: %v", err)
	}
	defer mod.Close(ctx)

	// Call add function
	addFn := mod.ExportedFunction("add")
	if addFn == nil {
		log.Fatal("add function not found")
	}

	results, err := addFn.Call(ctx, wasmtime.EncodeI32(5), wasmtime.EncodeI32(7))
	if err != nil {
		log.Fatalf("Failed to call add: %v", err)
	}

	fmt.Printf("5 + 7 = %d\n", wasmtime.DecodeI32(results[0]))

	// Call multiply function
	mulFn := mod.ExportedFunction("multiply")
	if mulFn == nil {
		log.Fatal("multiply function not found")
	}

	results, err = mulFn.Call(ctx, wasmtime.EncodeI32(6), wasmtime.EncodeI32(7))
	if err != nil {
		log.Fatalf("Failed to call multiply: %v", err)
	}

	fmt.Printf("6 * 7 = %d\n", wasmtime.DecodeI32(results[0]))
}
