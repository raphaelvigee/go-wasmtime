package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeU32(t *testing.T) {
	tests := []uint32{0, 1, 42, 0xFFFFFFFF}
	for _, want := range tests {
		encoded := EncodeU32(want)
		got := DecodeU32(encoded)
		assert.Equal(t, want, got, "EncodeU32/DecodeU32(%d) mismatch", want)
	}
}

func TestEncodeDecodeExternref(t *testing.T) {
	tests := []uintptr{0, 1, 0x1234, 0xDEADBEEF}
	for _, want := range tests {
		encoded := EncodeExternref(want)
		got := DecodeExternref(encoded)
		assert.Equal(t, want, got, "EncodeExternref/DecodeExternref(0x%x) mismatch", want)
	}
}

func TestGlobalAccess(t *testing.T) {

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	// WAT module with a global variable
	wat := `
	(module
		(global $counter (export "counter") (mut i32) (i32.const 0))
	)`

	compiled, err := r.CompileModule(t.Context(), []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(t.Context(), compiled)
	require.NoError(t, err)
	defer mod.Close(t.Context())

	// Get the global
	counter := mod.ExportedGlobal("counter")
	require.NotNil(t, counter, "counter global not found")

	// Test getting the initial value
	val := counter.Get(t.Context())
	assert.Equal(t, int32(0), DecodeI32(val), "Initial counter value")

	// Test setting a new value
	require.NoError(t, counter.Set(t.Context(), EncodeI32(42)))

	// Verify the new value
	val = counter.Get(t.Context())
	assert.Equal(t, int32(42), DecodeI32(val), "After set, counter value")
}

func TestTableAccess(t *testing.T) {

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	// WAT module with a table
	wat := `
	(module
		(table $tbl (export "tbl") 10 funcref)
	)`

	compiled, err := r.CompileModule(t.Context(), []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	mod, err := r.Instantiate(t.Context(), compiled)
	require.NoError(t, err)
	defer mod.Close(t.Context())

	// Get the table
	tbl := mod.ExportedTable("tbl")
	require.NotNil(t, tbl, "tbl table not found")

	// Test getting the size
	size := tbl.Size(t.Context())
	assert.Equal(t, uint32(10), size, "Table size")

	// Test growing the table
	prevSize, ok := tbl.Grow(t.Context(), 5)
	require.True(t, ok, "Table grow failed")
	assert.Equal(t, uint32(10), prevSize, "Previous size")

	// Verify new size
	newSize := tbl.Size(t.Context())
	assert.Equal(t, uint32(15), newSize, "New table size")
}
