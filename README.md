# purego-wasmtime

A Go library for executing WebAssembly using [wasmtime](https://github.com/bytecodealliance/wasmtime) **without CGO**, powered by [purego](https://github.com/ebitengine/purego).

## Features

- ✅ **Zero CGO** - Pure Go bindings using purego
- 📦 **Automatic Setup** - Downloads and caches wasmtime C library automatically
- 🚀 **WASI Support** - Full access to WASI features (filesystem, environment, stdio)
- 🎯 **Simple API** - Clean, idiomatic Go interface
- 🔒 **Type Safe** - Strong typing with Go's type system
- 🏗️ **Multi-Platform** - Supports macOS (aarch64/x86_64) and Linux (aarch64/x86_64)

## Installation

```bash
go get github.com/rvigee/purego-wasmtime
```

No additional setup required! The library will automatically download the appropriate wasmtime binary for your platform on first use.

## Quick Start

### Simple WAT Execution

```go
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
    r, _ := wasmtime.NewRuntime(ctx)
    defer r.Close(ctx)
    
    // Compile WAT to module
    wat := `
    (module
      (func (export "add") (param i32 i32) (result i32)
        local.get 0
        local.get 1
        i32.add
      )
    )`
    
    compiled, _ := r.CompileModule(ctx, []byte(wat))
    defer compiled.Close()
    
    // Instantiate
    mod, _ := r.Instantiate(ctx, compiled)
    defer mod.Close(ctx)
    
    // Get exported function
    addFn := mod.ExportedFunction("add")
    
    // Call function with encoded parameters
    results, _ := addFn.Call(ctx, wasmtime.EncodeI32(5), wasmtime.EncodeI32(7))
    
    // Decode result
    result := wasmtime.DecodeI32(results[0])
    fmt.Printf("5 + 7 = %d\n", result)
}
```

### WASI Example

```go
package main

import (
    "context"
    "os"
    
    wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
    ctx := context.Background()
    
    // Create runtime with WASI configuration
    config := wasmtime.NewRuntimeConfig().
        WithWASI(
            wasmtime.NewWASIConfig().
                WithInheritStdio().
                WithEnv("KEY", "value").
                WithArgs("program", "arg1"),
        )
    
    r, _ := wasmtime.NewRuntimeWithConfig(ctx, config)
    defer r.Close(ctx)
    
    // Load and compile WASI module
    wat, _ := os.ReadFile("program.wat")
    compiled, _ := r.CompileModule(ctx, wat)
    defer compiled.Close()
    
    // Instantiate with WASI support
    mod, _ := r.InstantiateWithWASI(ctx, compiled)
    defer mod.Close(ctx)
    
    // Call _start
    startFn := mod.ExportedFunction("_start")
    startFn.Call(ctx)
}
```

## API Overview

### Runtime

- `NewRuntime(ctx)` - Create a WebAssembly runtime
- `NewRuntimeWithConfig(ctx, config)` - Create runtime with configuration
- `runtime.CompileModule(ctx, binary)` - Compile WAT or WASM
- `runtime.Instantiate(ctx, compiled)` - Instantiate without WASI
- `runtime.InstantiateWithWASI(ctx, compiled)` - Instantiate with WASI
- `runtime.Close(ctx)` - Close and cleanup

### Module & Functions

- `module.ExportedFunction(name)` - Get an exported function
- `function.Call(ctx, params...)` - Call function with encoded parameters
- `module.Close(ctx)` - Close module

### Value Encoding/Decoding

- `EncodeI32(v)` / `DecodeI32(v)` - int32 values
- `EncodeI64(v)` / `DecodeI64(v)` - int64 values
- `EncodeF32(v)` / `DecodeF32(v)` - float32 values
- `EncodeF64(v)` / `DecodeF64(v)` - float64 values

### WASI Configuration

- `NewWASIConfig()` - Create WASI configuration
- `.WithArgs(args...)` - Set command-line arguments (variadic)
- `.WithEnv(key, value)` - Set single environment variable
- `.WithEnvs(map[string]string)` - Set multiple environment variables
- `.WithPreopenDir(host, guest)` - Grant directory access
- `.WithInheritStdio()` - Inherit stdin/stdout/stderr
- `.WithInheritArgs()` / `.WithInheritEnv()` - Inherit from host

### Runtime Configuration

- `NewRuntimeConfig()` - Create runtime configuration
- `.WithWASI(wasiConfig)` - Add WASI support

## Environment Variables

- `WASMTIME_LIB_PATH` - Override automatic download with custom wasmtime library path

## Platform Support

| Platform | Architecture | Status |
|----------|-------------|--------|
| macOS | x86_64 | ✅ Supported |
| macOS | aarch64 (Apple Silicon) | ✅ Supported |
| Linux | x86_64 | ✅ Supported |
| Linux | aarch64 | ✅ Supported |
| Windows | x86_64 | 🚧 Planned |

## How It Works

1. **Auto-Download**: On first use, the library detects your platform and downloads the appropriate wasmtime C API binary from GitHub releases (v40.0.0)
2. **Caching**: The library is extracted to `~/.local/share/purego-wasmtime/` and reused for subsequent runs
3. **Purego Bindings**: Uses purego to call wasmtime C functions without CGO
4. **Clean API**: Wraps low-level C calls in idiomatic Go interfaces

## Examples

See the `examples/` directory for complete programs:
- `simple.go` - Basic module compilation and instantiation
- `wasi_hello.go` - WASI program with environment and stdio

## Testing

Run tests with CGO explicitly disabled:

```bash
CGO_ENABLED=0 go test -v ./...
```

## License

MIT License - see LICENSE file for details

## Credits

- [wasmtime](https://github.com/bytecodealliance/wasmtime) - WebAssembly runtime
- [purego](https://github.com/ebitengine/purego) - C function calling without CGO
