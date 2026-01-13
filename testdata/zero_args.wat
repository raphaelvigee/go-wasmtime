(module
  (func (export "get_forty_two") (result i32)
    i32.const 42
  )
  (func (export "add_ten_twenty") (result i32)
    i32.const 10
    i32.const 20
    i32.add
  )
)
