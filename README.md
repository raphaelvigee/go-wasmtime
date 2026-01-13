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
    "log"
    wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
    // Create engine and store
    engine, _ := wasmtime.NewEngine()
    defer engine.Close()
    
    store, _ := wasmtime.NewStore(engine)
    defer store.Close()
    
    // Compile WAT to module
    wat := `
    (module
      (func (export "add") (param i32 i32) (result i32)
        local.get 0
        local.get 1
        i32.add
      )
    )`
    
    module, _ := wasmtime.NewModuleFromWAT(engine, wat)
    defer module.Close()
    
    // Instantiate
    instance, _ := wasmtime.NewInstance(store, module, nil)
    
    // Get exported function
    addFunc, _ := instance.GetExport("add")
    log.Printf("Function exported: %v", addFunc != nil)
}
```

### WASI Example

```go
package main

import (
    wasmtime "github.com/rvigee/purego-wasmtime"
)

func main() {
    engine, _ := wasmtime.NewEngine()
    defer engine.Close()
    
    store, _ := wasmtime.NewStore(engine)
    defer store.Close()
    
    // Configure WASI
    wasiConfig, _ := wasmtime.NewWASIConfig()
    wasiConfig.
        WithInheritStdio().
        WithEnv(map[string]string{"KEY": "value"}).
        WithArgs([]string{"program", "arg1"}).
        WithPreopenDir("/tmp", "/tmp")
    
    wasiConfig.Apply(store)
    
    // Load and run WASI module
    module, _ := wasmtime.NewModuleFromWATFile(engine, "program.wat")
    defer module.Close()
    
    instance, _ := wasmtime.NewInstance(store, module, nil)
    instance.Call("_start")
}
```

## API Overview

### Engine & Store

- `NewEngine()` - Create a WebAssembly engine
- `NewStore(engine)` - Create an execution store

### Modules

- `NewModuleFromWAT(engine, wat)` - Compile WAT text
- `NewModuleFromWATFile(engine, path)` - Load WAT from file
- `NewModuleFromWASM(engine, bytes)` - Compile WASM binary
- `NewModuleFromWASMFile(engine, path)` - Load WASM from file

### Instances

- `NewInstance(store, module, imports)` - Instantiate a module
- `instance.GetExport(name)` - Get exported function/memory
- `instance.Call(name, args...)` - Call exported function

### WASI Configuration

- `NewWASIConfig()` - Create WASI configuration
- `.WithArgs([]string)` - Set command-line arguments
- `.WithEnv(map[string]string)` - Set environment variables
- `.WithPreopenDir(host, guest)` - Grant directory access
- `.WithInheritStdio()` - Inherit stdin/stdout/stderr
- `.Apply(store)` - Apply configuration to store

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
