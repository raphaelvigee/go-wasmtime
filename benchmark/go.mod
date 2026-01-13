module github.com/rvigee/purego-wasmtime/benchmark

go 1.25.0

replace github.com/rvigee/purego-wasmtime => ../

require (
	github.com/bytecodealliance/wasmtime-go/v28 v28.0.0
	github.com/rvigee/purego-wasmtime v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.11.1
	github.com/tetratelabs/wazero v1.11.0
	github.com/wasmerio/wasmer-go v1.0.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	golang.org/x/sys v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
