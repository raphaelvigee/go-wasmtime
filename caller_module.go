package wasmtime

import (
	"context"

	"github.com/rvigee/purego-wasmtime/api"
)

// callerModule is a wrapper that provides access to the calling module's exports
// This is used for GoModuleFunc to access the calling WASM module's memory and other exports
type callerModule struct {
	caller   uintptr
	store    wasmtime_context_t
	bindings *bindings
}

func (cm *callerModule) Name() string {
	return "caller"
}

func (cm *callerModule) ExportedFunction(name string) api.Function {
	return nil
}

func (cm *callerModule) ExportedFunctionDefinitions() map[string]api.FunctionDefinition {
	return nil
}

func (cm *callerModule) ExportedMemory(name string) api.Memory {
	nameBytes := []byte(name + "\x00")
	var ext wasmtime_extern_t

	found := cm.bindings.wasmtime_caller_export_get(cm.caller, &nameBytes[0], uintptr(len(name)), &ext)
	if !found || ext.kind != WASMTIME_EXTERN_MEMORY {
		return nil
	}

	mem := ext.AsMemory()
	return &memory{
		val:      *mem,
		store:    0,
		storeCtx: cm.store,
		bindings: cm.bindings,
	}
}

func (cm *callerModule) ExportedGlobal(name string) api.Global {
	nameBytes := []byte(name + "\x00")
	var ext wasmtime_extern_t

	found := cm.bindings.wasmtime_caller_export_get(cm.caller, &nameBytes[0], uintptr(len(name)), &ext)
	if !found || ext.kind != WASMTIME_EXTERN_GLOBAL {
		return nil
	}

	glob := ext.AsGlobal()
	return &global{
		val:      *glob,
		store:    0,
		storeCtx: cm.store,
		bindings: cm.bindings,
	}
}

func (cm *callerModule) ExportedTable(name string) api.Table {
	nameBytes := []byte(name + "\x00")
	var ext wasmtime_extern_t

	found := cm.bindings.wasmtime_caller_export_get(cm.caller, &nameBytes[0], uintptr(len(name)), &ext)
	if !found || ext.kind != WASMTIME_EXTERN_TABLE {
		return nil
	}

	tbl := ext.AsTable()
	return &table{
		val:      *tbl,
		store:    0,
		storeCtx: cm.store,
		bindings: cm.bindings,
	}
}

func (cm *callerModule) Close(ctx context.Context) error {
	return nil
}
