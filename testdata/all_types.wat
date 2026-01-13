;; Comprehensive WebAssembly type tests
;; Tests all value types: i32, i64, f32, f64, and multi-value returns

(module
  ;; ========================================
  ;; i32 Tests
  ;; ========================================
  
  ;; Identity function for i32
  (func $i32_identity (param $a i32) (result i32)
    local.get $a
  )
  (export "i32_identity" (func $i32_identity))
  
  ;; Add two i32 values
  (func $i32_add (param $a i32) (param $b i32) (result i32)
    local.get $a
    local.get $b
    i32.add
  )
  (export "i32_add" (func $i32_add))
  
  ;; Multiply two i32 values
  (func $i32_multiply (param $a i32) (param $b i32) (result i32)
    local.get $a
    local.get $b
    i32.mul
  )
  (export "i32_multiply" (func $i32_multiply))
  
  ;; ========================================
  ;; i64 Tests
  ;; ========================================
  
  ;; Identity function for i64
  (func $i64_identity (param $a i64) (result i64)
    local.get $a
  )
  (export "i64_identity" (func $i64_identity))
  
  ;; Add two i64 values
  (func $i64_add (param $a i64) (param $b i64) (result i64)
    local.get $a
    local.get $b
    i64.add
  )
  (export "i64_add" (func $i64_add))
  
  ;; Multiply two i64 values
  (func $i64_multiply (param $a i64) (param $b i64) (result i64)
    local.get $a
    local.get $b
    i64.mul
  )
  (export "i64_multiply" (func $i64_multiply))
  
  ;; ========================================
  ;; f32 Tests
  ;; ========================================
  
  ;; Identity function for f32
  (func $f32_identity (param $a f32) (result f32)
    local.get $a
  )
  (export "f32_identity" (func $f32_identity))
  
  ;; Add two f32 values
  (func $f32_add (param $a f32) (param $b f32) (result f32)
    local.get $a
    local.get $b
    f32.add
  )
  (export "f32_add" (func $f32_add))
  
  ;; Multiply two f32 values
  (func $f32_multiply (param $a f32) (param $b f32) (result f32)
    local.get $a
    local.get $b
    f32.mul
  )
  (export "f32_multiply" (func $f32_multiply))
  
  ;; Square root of f32
  (func $f32_sqrt (param $a f32) (result f32)
    local.get $a
    f32.sqrt
  )
  (export "f32_sqrt" (func $f32_sqrt))
  
  ;; ========================================
  ;; f64 Tests
  ;; ========================================
  
  ;; Identity function for f64
  (func $f64_identity (param $a f64) (result f64)
    local.get $a
  )
  (export "f64_identity" (func $f64_identity))
  
  ;; Add two f64 values
  (func $f64_add (param $a f64) (param $b f64) (result f64)
    local.get $a
    local.get $b
    f64.add
  )
  (export "f64_add" (func $f64_add))
  
  ;; Multiply two f64 values
  (func $f64_multiply (param $a f64) (param $b f64) (result f64)
    local.get $a
    local.get $b
    f64.mul
  )
  (export "f64_multiply" (func $f64_multiply))
  
  ;; Square root of f64
  (func $f64_sqrt (param $a f64) (result f64)
    local.get $a
    f64.sqrt
  )
  (export "f64_sqrt" (func $f64_sqrt))
  
  ;; ========================================
  ;; Mixed Type Tests
  ;; ========================================
  
  ;; Function with multiple parameter types
  ;; Computes: (i32 + i64) * f32 + f64
  (func $mixed_params (param $i i32) (param $j i64) (param $f f32) (param $d f64) (result f64)
    ;; Convert i32 to f64
    local.get $i
    f64.convert_i32_s
    
    ;; Convert i64 to f64 and add
    local.get $j
    f64.convert_i64_s
    f64.add
    
    ;; Convert f32 to f64 and multiply
    local.get $f
    f64.promote_f32
    f64.mul
    
    ;; Add f64
    local.get $d
    f64.add
  )
  (export "mixed_params" (func $mixed_params))
  
  ;; Multi-value return with different types
  ;; Returns: (i32, f64, i64)
  (func $multi_return (result i32 f64 i64)
    i32.const 42
    f64.const 3.14159
    i64.const 9223372036854775807
  )
  (export "multi_return" (func $multi_return))
  
  ;; Multi-value return with all basic types
  ;; Returns: (i32, i64, f32, f64)
  (func $all_types_return (result i32 i64 f32 f64)
    i32.const 100
    i64.const 200
    f32.const 1.5
    f64.const 2.5
  )
  (export "all_types_return" (func $all_types_return))
  
  ;; Function that converts between types
  ;; Takes i32, returns as f64
  (func $i32_to_f64 (param $a i32) (result f64)
    local.get $a
    f64.convert_i32_s
  )
  (export "i32_to_f64" (func $i32_to_f64))
  
  ;; Function that converts f64 to i32 (truncate)
  (func $f64_to_i32 (param $a f64) (result i32)
    local.get $a
    i32.trunc_f64_s
  )
  (export "f64_to_i32" (func $f64_to_i32))
)
