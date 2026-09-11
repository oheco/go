// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// AArch64 TLSDESC resolvers preserve all registers except x0 and NZCV.
// Preserve the scratch register used to call the resolver and the flags too:
// runtime.load_g/save_g promise to clobber only R0 and R27.
// musl allocates dynamic TLS for existing threads at dlopen, and for new
// threads before they start; these accesses need no allocation in a signal.
.text
.p2align 2
.globl x_cgo_ohos_load_g
.hidden x_cgo_ohos_load_g
.type x_cgo_ohos_load_g,%function
x_cgo_ohos_load_g:
    stp x1, x30, [sp, #-32]!
    mrs x1, nzcv
    str x1, [sp, #16]
    adrp x0, :tlsdesc:x_cgo_ohos_g
    ldr x1, [x0, :tlsdesc_lo12:x_cgo_ohos_g]
    add x0, x0, :tlsdesc_lo12:x_cgo_ohos_g
    .tlsdesccall x_cgo_ohos_g
    blr x1
    mrs x1, tpidr_el0
    ldr x0, [x1, x0]
    ldr x1, [sp, #16]
    msr nzcv, x1
    ldp x1, x30, [sp], #32
    ret
.size x_cgo_ohos_load_g, .-x_cgo_ohos_load_g

.p2align 2
.globl x_cgo_ohos_save_g
.hidden x_cgo_ohos_save_g
.type x_cgo_ohos_save_g,%function
x_cgo_ohos_save_g:
    stp x1, x30, [sp, #-48]!
    stp x0, x2, [sp, #16]
    mrs x1, nzcv
    str x1, [sp, #32]
    adrp x0, :tlsdesc:x_cgo_ohos_g
    ldr x1, [x0, :tlsdesc_lo12:x_cgo_ohos_g]
    add x0, x0, :tlsdesc_lo12:x_cgo_ohos_g
    .tlsdesccall x_cgo_ohos_g
    blr x1
    mrs x1, tpidr_el0
    ldr x2, [sp, #16]
    str x2, [x1, x0]
    ldr x1, [sp, #32]
    msr nzcv, x1
    ldp x0, x2, [sp, #16]
    ldp x1, x30, [sp], #48
    ret
.size x_cgo_ohos_save_g, .-x_cgo_ohos_save_g

.section .tbss,"awT",%nobits
.p2align 3
.type x_cgo_ohos_g,%tls_object
.size x_cgo_ohos_g,8
x_cgo_ohos_g:
    .zero 8

.section .note.GNU-stack,"",%progbits
