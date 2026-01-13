(module
  (memory (export "memory") 1)
  
  ;; Reverse a string
  ;; Params: ptr (i32), len (i32)
  ;; Results: ptr (i32), len (i32)
  ;; Writes result to offset 1024
  (func (export "reverse") (param $ptr i32) (param $len i32) (result i32 i32)
    (local $i i32)
    (local $j i32)
    (local $char i32)
    
    ;; $i = 0
    (local.set $i (i32.const 0))
    ;; $j = $len - 1
    (local.set $j (local.get $len))
    (local.set $j (i32.sub (local.get $j) (i32.const 1)))
    
    (block $break
      (loop $loop
        ;; if i >= len, break
        (br_if $break (i32.ge_u (local.get $i) (local.get $len)))
        
        ;; load char from input at ptr + i
        (local.set $char (i32.load8_u (i32.add (local.get $ptr) (local.get $i))))
        
        ;; store char to output at 1024 + j
        (i32.store8 (i32.add (i32.const 1024) (local.get $j)) (local.get $char))
        
        ;; i++
        (local.set $i (i32.add (local.get $i) (i32.const 1)))
        ;; j--
        (local.set $j (i32.sub (local.get $j) (i32.const 1)))
        
        (br $loop)
      )
    )
    
    ;; return 1024, len
    (i32.const 1024)
    (local.get $len)
  )
)
