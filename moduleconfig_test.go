package wasmtime

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleConfig(t *testing.T) {
	require.NoError(t, Initialize())

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	// WAT module that prints to stdout
	wat := `
	(module
		(import "wasi_snapshot_preview1" "fd_write"
			(func $fd_write (param i32 i32 i32 i32) (result i32)))
		
		(memory 1)
		(export "memory" (memory 0))
		
		(data (i32.const 0) "Hello from config!\n")
		
		(func (export "_start")
			;; iovec struct at offset 100
			(i32.store (i32.const 100) (i32.const 0))   ;; buf pointer
			(i32.store (i32.const 104) (i32.const 19))  ;; buf length
			
			;; Call fd_write(1, 100, 1, 200)
			(call $fd_write
				(i32.const 1)    ;; stdout
				(i32.const 100)  ;; iovs
				(i32.const 1)    ;; iovs_len
				(i32.const 200)) ;; nwritten
			drop
		)
	)`

	compiled, err := r.CompileModule(t.Context(), []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	t.Run("custom_stdout", func(t *testing.T) {
		var stdout bytes.Buffer

		config := NewModuleConfig().
			WithName("test_module").
			WithStdout(&stdout)

		require.NotNil(t, config)

		mc := config.(*moduleConfig)
		assert.Equal(t, "test_module", mc.name)
		assert.Equal(t, &stdout, mc.stdout)
	})

	t.Run("args_and_env", func(t *testing.T) {
		config := NewModuleConfig().
			WithArgs("program", "arg1", "arg2").
			WithEnv("KEY1", "value1").
			WithEnv("KEY2", "value2").
			WithEnvs(map[string]string{
				"KEY3": "value3",
				"KEY4": "value4",
			})

		mc := config.(*moduleConfig)

		assert.Len(t, mc.args, 3)
		assert.Equal(t, []string{"program", "arg1", "arg2"}, mc.args)

		assert.Len(t, mc.env, 4)
		assert.Equal(t, "value1", mc.env["KEY1"])
		assert.Equal(t, "value3", mc.env["KEY3"])
	})

	t.Run("dir_preopens", func(t *testing.T) {
		config := NewModuleConfig().
			WithDirPreopen("/host/path1", "/guest/path1").
			WithDirPreopen("/host/path2", "/guest/path2")

		mc := config.(*moduleConfig)

		assert.Len(t, mc.preopens, 2)
		assert.Equal(t, "/host/path1", mc.preopens["/guest/path1"])
	})

	t.Run("start_functions", func(t *testing.T) {
		config := NewModuleConfig().
			WithStartFunctions("_initialize", "_start")

		mc := config.(*moduleConfig)

		assert.Len(t, mc.startFunctions, 2)
		assert.Equal(t, []string{"_initialize", "_start"}, mc.startFunctions)
	})

	t.Run("stdin_capture", func(t *testing.T) {
		input := "test input\n"
		stdin := strings.NewReader(input)

		config := NewModuleConfig().
			WithStdin(stdin)

		mc := config.(*moduleConfig)
		assert.Equal(t, stdin, mc.stdin)
	})

	t.Run("chaining", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("input")

		config := NewModuleConfig().
			WithName("chained").
			WithArgs("prog", "a1").
			WithEnv("K", "V").
			WithStdin(stdin).
			WithStdout(&stdout).
			WithStderr(&stderr).
			WithDirPreopen("/h", "/g").
			WithStartFunctions("_start")

		mc := config.(*moduleConfig)

		assert.Equal(t, "chained", mc.name)
		assert.Len(t, mc.args, 2)
		assert.Len(t, mc.env, 1)
		assert.NotNil(t, mc.stdin)
		assert.NotNil(t, mc.stdout)
		assert.NotNil(t, mc.stderr)
		assert.Len(t, mc.preopens, 1)
		assert.Len(t, mc.startFunctions, 1)
	})
}

func TestCompiledModuleIntrospection(t *testing.T) {
	require.NoError(t, Initialize())

	r, err := NewRuntime(t.Context())
	require.NoError(t, err)
	defer r.Close(t.Context())

	wat := `
	(module
		(func (export "add") (param i32 i32) (result i32)
			local.get 0
			local.get 1
			i32.add
		)
		(memory (export "memory") 1)
		(global (export "counter") (mut i32) (i32.const 0))
	)`

	compiled, err := r.CompileModule(t.Context(), []byte(wat))
	require.NoError(t, err)
	defer compiled.Close()

	t.Run("exported_functions", func(t *testing.T) {
		funcs := compiled.ExportedFunctions()
		assert.NotNil(t, funcs)
	})

	t.Run("exported_memories", func(t *testing.T) {
		mems := compiled.ExportedMemories()
		assert.NotNil(t, mems)
	})

	t.Run("custom_sections", func(t *testing.T) {
		sections := compiled.CustomSections()
		// nil is acceptable for no custom sections
		_ = sections
	})

	t.Run("module_name", func(t *testing.T) {
		name := compiled.Name()
		// Empty string is acceptable for unnamed modules
		_ = name
	})
}

func TestMemoryDefinition(t *testing.T) {
	tests := []struct {
		name       string
		min        uint32
		max        uint32
		maxEncoded bool
	}{
		{"min_only", 1, 0, false},
		{"with_max", 1, 10, true},
		{"large_min", 100, 1000, true},
		{"zero_min", 0, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := newMemoryDefinition(tt.min, tt.max, tt.maxEncoded)

			assert.Equal(t, tt.min, md.Min())
			assert.Equal(t, tt.max, md.Max())
			assert.Equal(t, tt.maxEncoded, md.IsMaxEncoded())
		})
	}
}
