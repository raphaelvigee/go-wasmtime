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
		log.Fatal(err)
	}
	defer r.Close(ctx)

	// Compile WAT module with globals and tables
	wat := `
	(module
		;; Mutable global counter
		(global $counter (export "counter") (mut i32) (i32.const 0))
		
		;; Immutable constant
		(global $pi (export "pi") f32 (f32.const 3.14159))
		
		;; Exported table
		(table $dispatch (export "dispatch") 5 funcref)
		
		;; Function to increment counter
		(func $increment (export "increment")
			global.get $counter
			i32.const 1
			i32.add
			global.set $counter
		)
		
		;; Function to get counter value
		(func $get_count (export "get_count") (result i32)
			global.get $counter
		)
	)`

	compiled, err := r.CompileModule(ctx, []byte(wat))
	if err != nil {
		log.Fatal(err)
	}
	defer compiled.Close()

	// Instantiate
	mod, err := r.Instantiate(ctx, compiled)
	if err != nil {
		log.Fatal(err)
	}
	defer mod.Close(ctx)

	fmt.Println("=== Global Variables Demo ===")

	// Access mutable global
	counter := mod.ExportedGlobal("counter")
	if counter != nil {
		fmt.Printf("Initial counter: %d\n", wasmtime.DecodeI32(counter.Get(ctx)))

		// Set counter to 10
		if err := counter.Set(ctx, wasmtime.EncodeI32(10)); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("After set: %d\n", wasmtime.DecodeI32(counter.Get(ctx)))

		// Call increment function
		incFn := mod.ExportedFunction("increment")
		if _, err := incFn.Call(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("After increment: %d\n", wasmtime.DecodeI32(counter.Get(ctx)))
	}

	// Access immutable global
	pi := mod.ExportedGlobal("pi")
	if pi != nil {
		fmt.Printf("Pi constant: %.5f\n", wasmtime.DecodeF32(pi.Get(ctx)))

		// Try to set immutable global (should fail)
		if err := pi.Set(ctx, wasmtime.EncodeF32(3.0)); err != nil {
			fmt.Printf("Expected error setting immutable global: %v\n", err)
		}
	}

	fmt.Println("\n=== Table Demo ===")

	// Access table
	dispatch := mod.ExportedTable("dispatch")
	if dispatch != nil {
		size := dispatch.Size(ctx)
		fmt.Printf("Table size: %d\n", size)

		// Grow the table
		prevSize, ok := dispatch.Grow(ctx, 3)
		if ok {
			fmt.Printf("Grew table from %d to %d elements\n", prevSize, dispatch.Size(ctx))
		} else {
			fmt.Println("Failed to grow table")
		}
	}

	fmt.Println("\n=== Memory Access Demo ===")

	// Access memory (if module had one)
	mem := mod.ExportedMemory("memory")
	if mem != nil {
		fmt.Printf("Memory size: %d pages\n", mem.Size(ctx))
		fmt.Printf("Memory data size: %d bytes\n", mem.DataSize(ctx))
	} else {
		fmt.Println("Module has no exported memory")
	}

	fmt.Println("\n=== Encoding Demo ===")

	// Demonstrate all encoding types
	fmt.Printf("U32: %d -> %d\n", uint32(42), wasmtime.DecodeU32(wasmtime.EncodeU32(42)))
	fmt.Printf("I32: %d -> %d\n", int32(-42), wasmtime.DecodeI32(wasmtime.EncodeI32(-42)))
	fmt.Printf("I64: %d -> %d\n", int64(9876543210), wasmtime.DecodeI64(wasmtime.EncodeI64(9876543210)))
	fmt.Printf("F32: %.2f -> %.2f\n", float32(3.14), wasmtime.DecodeF32(wasmtime.EncodeF32(3.14)))
	fmt.Printf("F64: %.5f -> %.5f\n", 2.71828, wasmtime.DecodeF64(wasmtime.EncodeF64(2.71828)))

	// Externref example
	ptr := uintptr(0xDEADBEEF)
	fmt.Printf("Externref: 0x%X -> 0x%X\n", ptr, wasmtime.DecodeExternref(wasmtime.EncodeExternref(ptr)))
}
