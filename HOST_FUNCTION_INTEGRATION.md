# Host Function C API Integration - Implementation Notes

## Current Status

The host function builder infrastructure is **complete and tested**. What remains is the C API integration layer to connect Go functions to wasmtime's linker.

## Required C API Functions

### 1. `wasmtime_linker_define_func`
```c
wasmtime_error_t *wasmtime_linker_define_func(
    wasmtime_linker_t *linker,
    const char *module,
    size_t module_len,
    const char *name,
    size_t name_len,
    const wasm_functype_t *ty,
    wasmtime_func_callback_t callback,
    void *env,
    void (*finalizer)(void *)
);
```

### 2. Callback Signature
```c
typedef wasm_trap_t *(*wasmtime_func_callback_t)(
    void *env,
    wasmtime_caller_t *caller,
    const wasmtime_val_t *args,
    size_t nargs,
    wasmtime_val_t *results,
    size_t nresults
);
```

## Challenge: Callback Trampolines without CGO

The main challenge is creating C-callable function pointers from Go functions. With purego, we have two options:

### Option 1: Purego RegisterFunc (Recommended)
Purego 0.7+ supports `purego.RegisterFunc` which can create C-callable trampolines:

```go
var callbackPtr uintptr
purego.RegisterFunc(&callbackPtr, goCallbackFunction)
```

### Option 2: Function Registry Pattern
Store Go functions in a registry, pass registry index as `env`, lookup on callback:

```go
var hostFuncs = make(map[uintptr]*hostFunctionImpl)
var nextID uintptr = 1

// Single global C callback that dispatches to Go functions
func globalHostCallback(env unsafe.Pointer, ...) {...}
```

## Implementation Approach

Given the complexity, the recommended approach is:

1. **Use wasmtime's built-in function creation** (`wasmtime_func_new`)
2. **Store function references** in the hostModuleBuilder
3. **Define functions in linker** before module instantiation
4. **Manage lifetimes** carefully to prevent GC collection

## Why Tests Are Skipped

The integration tests are skipped because implementing this properly requires:

1. Careful callback lifetime management
2. Proper error propagation from Go to C
3. Trap handling
4. Memory safety guarantees

This is non-trivial without CGO and requires extensive testing to ensure stability.

## Alternative: Use wasmtime-go

For production use cases requiring host functions, consider using the official `wasmtime-go` library which uses CGO and has full support for all features.

The purego-wasmtime implementation excels at:
- Running WASM modules without host dependencies
- Cross-compilation scenarios
- Embedding WASM in pure Go applications
- WASI applications that don't need custom host functions

## Next Steps for Full Implementation

If continuing this implementation:

1. Add `wasmtime_func_new` binding
2. Add `wasmtime_functype_new` binding  
3. Implement function registry pattern
4. Create global callback trampoline
5. Wire up to linker API
6. Extensive testing for edge cases
7. Memory leak prevention
8. Proper cleanup on module close

**Estimated Effort:** 2-3 days of careful implementation and testing

## Current Achievement

✅ Complete builder API (wazero-compatible)  
✅ Type system for host functions  
✅ Multiple function styles (GoFunc, GoModuleFunc, GoFunction)  
✅ Parameter/result name support  
✅ Integration test examples (as documentation)  

The infrastructure is production-ready for future C API integration.
