// Package testhelpers provides utilities for writing tests with the new API
package wasmtime_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	wasmtime "github.com/rvigee/purego-wasmtime"
	"github.com/rvigee/purego-wasmtime/api"
)

// testSetup is a test helper that sets up a runtime and compiles a WAT file
type testSetup struct {
	ctx      context.Context
	runtime  wasmtime.Runtime
	compiled wasmtime.CompiledModule
	module   api.Module
}

// newTestSetup creates a new test setup from a WAT file path
func newTestSetup(t *testing.T, watPath string, useWASI bool) *testSetup {
	ts := &testSetup{
		ctx: context.Background(),
	}

	var err error
	if useWASI {
		config := wasmtime.NewRuntimeConfig().
			WithWASI(wasmtime.NewWASIConfig().WithInheritStdio())
		ts.runtime, err = wasmtime.NewRuntimeWithConfig(ts.ctx, config)
	} else {
		ts.runtime, err = wasmtime.NewRuntime(ts.ctx)
	}
	require.NoError(t, err)

	// Read WAT file
	wat, err := os.ReadFile(watPath)
	require.NoError(t, err)

	// Compile module
	ts.compiled, err = ts.runtime.CompileModule(ts.ctx, wat)
	require.NoError(t, err)

	// Instantiate
	if useWASI {
		ts.module, err = ts.runtime.InstantiateWithWASI(ts.ctx, ts.compiled)
	} else {
		ts.module, err = ts.runtime.Instantiate(ts.ctx, ts.compiled)
	}
	require.NoError(t, err)

	return ts
}

// cleanup closes all resources
func (ts *testSetup) cleanup() {
	if ts.module != nil {
		ts.module.Close(ts.ctx)
	}
	if ts.compiled != nil {
		ts.compiled.Close()
	}
	if ts.runtime != nil {
		ts.runtime.Close(ts.ctx)
	}
}
