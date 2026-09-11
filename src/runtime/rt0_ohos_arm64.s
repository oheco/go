// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ohos

#include "textflag.h"

// HarmonyOS uses the same initial stack layout as the Linux ARM64 ABI.
TEXT _rt0_arm64_ohos(SB),NOSPLIT,$0
	JMP	_rt0_arm64(SB)

// Initialize the runtime when loaded by a C executable or dynamic loader.
TEXT _rt0_arm64_ohos_lib(SB),NOSPLIT,$0
	// musl invokes ELF constructors without argc/argv. Supply an empty,
	// terminated argument/environment/auxiliary vector, as Android does.
	// sysargs obtains the real auxiliary vector from /proc/self/auxv.
	MOVD	$0, R0
	MOVD	$_rt0_arm64_ohos_argv(SB), R1
	JMP	_rt0_arm64_lib(SB)

DATA _rt0_arm64_ohos_argv+0x00(SB)/8, $0 // end argv
DATA _rt0_arm64_ohos_argv+0x08(SB)/8, $0 // end envv
DATA _rt0_arm64_ohos_argv+0x10(SB)/8, $0 // end auxv
GLOBL _rt0_arm64_ohos_argv(SB), NOPTR, $0x18
