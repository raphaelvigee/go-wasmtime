package wasmtime

import "github.com/rvigee/purego-wasmtime/api"

// memoryDefinition implements api.MemoryDefinition
type memoryDefinition struct {
	min        uint32
	max        uint32
	maxEncoded bool
}

func (md *memoryDefinition) Min() uint32 {
	return md.min
}

func (md *memoryDefinition) Max() uint32 {
	return md.max
}

func (md *memoryDefinition) IsMaxEncoded() bool {
	return md.maxEncoded
}

// newMemoryDefinition creates a memory definition with the given limits.
func newMemoryDefinition(min, max uint32, maxEncoded bool) api.MemoryDefinition {
	return &memoryDefinition{
		min:        min,
		max:        max,
		maxEncoded: maxEncoded,
	}
}
