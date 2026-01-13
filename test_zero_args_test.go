package wasmtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
)

func TestZeroArgs(t *testing.T) {
	ts := newTestSetup(t, "testdata/zero_args.wat", false)
	defer ts.cleanup()

	fn := ts.module.ExportedFunction("get_forty_two")
	require.NotNil(t, fn, "Function get_forty_two should be exported")

	results, err := fn.Call(ts.ctx)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, int32(42), wasmtime.DecodeI32(results[0]))
}
