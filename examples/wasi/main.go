package main

import (
	"context"
	"fmt"
	"log"

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

	// Define the WASI module
	wat := `
	(module
	  (import "wasi_snapshot_preview1" "proc_exit" (func $proc_exit (param i32)))
	  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
	  
	  (memory (export "memory") 1)
	  
	  ;; Write "Hello from WASI!\n" to stdout
	  (data (i32.const 8) "Hello from WASI!\n")
	  
	  (func (export "_start")
	    ;; Set up iovec for fd_write
	    (i32.store (i32.const 0) (i32.const 8))   ;; iovec[0].buf = 8 (pointer to string)
	    (i32.store (i32.const 4) (i32.const 17))  ;; iovec[0].len = 17 (length of string)
	    
	    ;; Call fd_write(stdout=1, iovs=0, iovs_len=1, nwritten=100)
	    (drop
	      (call $fd_write
	        (i32.const 1)   ;; stdout
	        (i32.const 0)   ;; pointer to iovec
	        (i32.const 1)   ;; number of iovecs
	        (i32.const 100) ;; where to write number of bytes written
	      )
	    )
	    
	    ;; Exit with code 0
	    (call $proc_exit (i32.const 0))
	  )
	)
	`

	compiled, err := r.CompileModule(ctx, []byte(wat))
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
