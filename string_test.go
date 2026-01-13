package wasmtime_test

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringManipulation(t *testing.T) {
	ts := newTestSetup(t, "testdata/string.wat", false)
	defer ts.cleanup()

	// Get memory export
	mem := ts.module.ExportedMemory("memory")
	require.NotNil(t, mem)

	// Check initial size
	assert.GreaterOrEqual(t, mem.Size(ts.ctx), uint64(1))
	assert.GreaterOrEqual(t, mem.DataSize(ts.ctx), uintptr(65536))

	// Prepare string input
	input := "Hello, World!"
	inputBytes := []byte(input)
	inputPtr := uint64(0)
	inputLen := uint64(len(inputBytes))

	// Write string to memory
	dataPtr := mem.Data(ts.ctx)
	data := unsafe.Slice((*byte)(dataPtr), mem.DataSize(ts.ctx))
	copy(data[inputPtr:], inputBytes)

	// Call reverse function
	reverseFn := ts.module.ExportedFunction("reverse")
	require.NotNil(t, reverseFn)

	results, err := reverseFn.Call(ts.ctx, inputPtr, inputLen)
	require.NoError(t, err)
	require.Len(t, results, 2)

	resultPtr := results[0]
	resultLen := results[1]

	// Verify result length matches input length
	assert.Equal(t, inputLen, resultLen)
	// Verify result pointer is at 1024
	assert.Equal(t, uint64(1024), resultPtr)

	// Read result from memory
	// Need to refresh data pointer in case memory grew (though it shouldn't have here)
	dataPtr = mem.Data(ts.ctx)
	data = unsafe.Slice((*byte)(dataPtr), mem.DataSize(ts.ctx))
	resultBytes := data[resultPtr : resultPtr+resultLen]
	result := string(resultBytes)

	assert.Equal(t, "!dlroW ,olleH", result)
}
