package wasmtime

import (
	"unsafe"
)

// Opaque C types - these are pointers to C structs we don't need to know the internals of
type (
	wasm_engine_t      uintptr
	wasmtime_store_t   uintptr
	wasmtime_context_t uintptr
	wasmtime_module_t  uintptr
	wasmtime_error_t   uintptr
	wasm_trap_t        uintptr
	wasm_functype_t    uintptr
	wasi_config_t      uintptr
)

// wasmtime_func_t - From C header: struct with store_id and __private pointer
type wasmtime_func_t struct {
	store_id  uint64
	__private uintptr
}

// wasmtime_table_t - From C header
type wasmtime_table_t struct {
	store_id   uint64
	__private1 uint32
	__private2 uint32
}

// wasmtime_memory_t - From C header
type wasmtime_memory_t struct {
	store_id   uint64
	__private1 uint32
	__private2 uint32
}

// wasmtime_global_t - From C header
type wasmtime_global_t struct {
	store_id   uint64
	__private1 uint32
	__private2 uint32
	__private3 uint32
}

// wasmtime_instance_t - From C header: struct with store_id and __private size_t
//
//	typedef struct wasmtime_instance {
//	  uint64_t store_id;
//	  size_t __private;
//	} wasmtime_instance_t;
type wasmtime_instance_t struct {
	store_id  uint64
	__private uintptr // size_t in C
}

// wasm_byte_vec_t represents a vector of bytes in C
type wasm_byte_vec_t struct {
	size uintptr
	data *byte
}

// wasmtime_val_raw represents a raw WebAssembly value
// This is a union in C. The largest member is wasmtime_anyref_t (24 bytes),
// NOT v128 (16 bytes) as I initially thought!
type wasmtime_val_raw struct {
	data [24]byte // Union size must match C: sizeof(wasmtime_valunion_t) = 24
}

// wasmtime_val_t represents a WebAssembly value with its type
// The C struct has padding for alignment
// Total size: 1 + 7 + 24 = 32 bytes
type wasmtime_val_t struct {
	kind     uint8            // 1 byte at offset 0
	_padding [7]byte          // 7 bytes padding at offset 1-7
	of       wasmtime_val_raw // 24 bytes at offset 8
}

// Helper to get i32 from wasmtime_val_t
func (v *wasmtime_val_t) GetI32() int32 {
	return *(*int32)(unsafe.Pointer(&v.of.data[0]))
}

// Helper to set i32 in wasmtime_val_t
func (v *wasmtime_val_t) SetI32(val int32) {
	v.kind = 0
	*(*int32)(unsafe.Pointer(&v.of.data[0])) = val
}

// Helper to get i64 from wasmtime_val_t
func (v *wasmtime_val_t) GetI64() int64 {
	return *(*int64)(unsafe.Pointer(&v.of.data[0]))
}

// Helper to set i64 in wasmtime_val_t
func (v *wasmtime_val_t) SetI64(val int64) {
	v.kind = 1
	*(*int64)(unsafe.Pointer(&v.of.data[0])) = val
}

// wasmtime_extern_kind_t enum values
const (
	WASMTIME_EXTERN_FUNC   = 0
	WASMTIME_EXTERN_GLOBAL = 1
	WASMTIME_EXTERN_TABLE  = 2
	WASMTIME_EXTERN_MEMORY = 3
)

// wasmtime_extern_union is a union type for external items
// In C this is a union, but in Go with purego we represent it as raw bytes
// The size must be sizeof(largest_member):
// - wasmtime_func_t: 16 bytes (uint64 + uintptr)
// - wasmtime_table_t: 16 bytes
// - wasmtime_memory_t: 16 bytes
// - wasmtime_global_t: 20 bytes (uint64 + 3*uint32)
// So we need 20 bytes, but let's use 24 for alignment (8-byte aligned)
type wasmtime_extern_union struct {
	data [24]byte // Raw union data - reinterpret as needed
}

// wasmtime_extern_t represents an external item (function, memory, etc.)
// The C compiler adds automatic padding to align the union (which contains uint64).
// Even though the C header shows them adjacent, the actual memory layout has padding.
type wasmtime_extern_t struct {
	kind     wasmtime_extern_kind_t // uint8 at offset 0
	_padding [7]byte                // CRITICAL: Padding at offset 1-7 for alignment
	of       wasmtime_extern_union  // Union at offset 8
}

// Helper to get func from extern
func (e *wasmtime_extern_t) AsFunc() *wasmtime_func_t {
	return (*wasmtime_func_t)(unsafe.Pointer(&e.of.data[0]))
}

// Helper to get table from extern
func (e *wasmtime_extern_t) AsTable() *wasmtime_table_t {
	return (*wasmtime_table_t)(unsafe.Pointer(&e.of.data[0]))
}

// Helper to get memory from extern
func (e *wasmtime_extern_t) AsMemory() *wasmtime_memory_t {
	return (*wasmtime_memory_t)(unsafe.Pointer(&e.of.data[0]))
}

// Helper to get global from extern
func (e *wasmtime_extern_t) AsGlobal() *wasmtime_global_t {
	return (*wasmtime_global_t)(unsafe.Pointer(&e.of.data[0]))
}

// wasmtime_extern_kind_t is an alias for uint8
type wasmtime_extern_kind_t = uint8

// WASI permission flags
const (
	WASI_DIR_PERMS_READ   = 1
	WASI_DIR_PERMS_WRITE  = 2
	WASI_FILE_PERMS_READ  = 1
	WASI_FILE_PERMS_WRITE = 2
)

// Helper functions for creating byte vectors

// newByteVec creates a new wasm_byte_vec_t from a Go byte slice
func newByteVec(data []byte) wasm_byte_vec_t {
	if len(data) == 0 {
		return wasm_byte_vec_t{size: 0, data: nil}
	}
	return wasm_byte_vec_t{
		size: uintptr(len(data)),
		data: &data[0],
	}
}

// toGoBytes converts a wasm_byte_vec_t to a Go byte slice
func (v *wasm_byte_vec_t) toGoBytes() []byte {
	if v.size == 0 || v.data == nil {
		return nil
	}
	return unsafe.Slice(v.data, v.size)
}

// Helper functions for creating C strings

// cString creates a null-terminated C string from a Go string
func cString(s string) *byte {
	b := append([]byte(s), 0)
	return &b[0]
}

// cStringArray creates a NULL-terminated array of C strings
func cStringArray(strs []string) **byte {
	if len(strs) == 0 {
		return nil
	}
	ptrs := make([]*byte, len(strs)+1)
	for i, s := range strs {
		ptrs[i] = cString(s)
	}
	ptrs[len(strs)] = nil
	return &ptrs[0]
}
