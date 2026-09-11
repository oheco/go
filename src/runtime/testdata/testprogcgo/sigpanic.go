// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// This program will crash.
// We want to test unwinding from sigpanic into C code (without a C symbolizer).

/*
#cgo CFLAGS: -O0

#if defined(__OHOS__)
#include <signal.h>
#include <stdlib.h>

static void __attribute__((constructor)) defaultSigsegv(void) {
	// The OHOS loader preinstalls a DFX crash handler. This test needs
	// the default disposition before Go starts, so it tests Go's C
	// traceback instead of forwarding the fault to that native handler.
	if (getenv("GO_TEST_CGO_DEFAULT_SIGSEGV") != NULL)
		signal(SIGSEGV, SIG_DFL);
}
#endif

char *pnil;

static int f1(void) {
	*pnil = 0;
	return 0;
}
*/
import "C"

func init() {
	register("TracebackSigpanic", TracebackSigpanic)
}

func TracebackSigpanic() {
	C.f1()
}
