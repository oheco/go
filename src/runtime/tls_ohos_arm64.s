// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// The musl loader cannot allocate initial-exec TLS for dlopen'ed Go libraries.
// The cgo veneers use AArch64 TLSDESC and preserve every register except R0.
// As on other ARM64 ports, these helpers may clobber R0 and R27.
TEXT runtime·load_g(SB),NOSPLIT,$0
	MOVB runtime·iscgo(SB), R0
	CBZ R0, done
	MOVD _cgo_ohos_load_g(SB), R27
	CALL (R27)
	MOVD R0, g
done:
	RET

TEXT runtime·save_g(SB),NOSPLIT,$0
	MOVB runtime·iscgo(SB), R0
	CBZ R0, done
	MOVD _cgo_ohos_save_g(SB), R27
	MOVD g, R0
	CALL (R27)
done:
	RET
