(module
  ;; 0 parameters
  (func (export "constant") (result i32)
    i32.const 100
  )
  
  ;; 1 parameter
  (func (export "negate") (param i32) (result i32)
    local.get 0
    i32.const -1
    i32.mul
  )
  
  (func (export "double") (param i32) (result i32)
    local.get 0
    i32.const 2
    i32.mul
  )
  
  (func (export "square") (param i32) (result i32)
    local.get 0
    local.get 0
    i32.mul
  )
  
  ;; 2 parameters
  (func (export "subtract") (param i32 i32) (result i32)
    local.get 0
    local.get 1
    i32.sub
  )
  
  (func (export "max") (param i32 i32) (result i32)
    local.get 0
    local.get 1
    local.get 0
    local.get 1
    i32.gt_s
    select
  )
)
