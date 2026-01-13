(module
  (import "wasi_snapshot_preview1" "proc_exit" (func $proc_exit (param i32)))
  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
  
  (memory (export "memory") 1)
  
  ;; Write "Hello from WASI!\n" to stdout
  (data (i32.const 8) "Hello from WASI!\n")
  
  (func (export "_start")
    ;; Set up iovec for fd_write
    (i32.store (i32.const 0) (i32.const 8))   ;; iovec[0].buf = 8 (pointer to string)
    (i32.store (i32.const 4) (i32.const 17))  ;; iovec[0].len = 17 (length of string)
    
    ;; Call fd_write(stdout=1, iovs=0, iovs_len=1, nwritten=100)
    (drop
      (call $fd_write
        (i32.const 1)   ;; stdout
        (i32.const 0)   ;; pointer to iovec
        (i32.const 1)   ;; number of iovecs
        (i32.const 100) ;; where to write number of bytes written
      )
    )
    
    ;; Exit with code 0
    (call $proc_exit (i32.const 0))
  )
)
