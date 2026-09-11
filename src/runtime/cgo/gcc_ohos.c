// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "libcgo.h"

extern char **environ;
char **x_cgo_ohos_environ;

// musl does not pass argv/envp to ELF constructors. Capture libc's
// environment during runtime startup for Go embedded in a C process.
static void inittls(void **tlsg, void **tlsbase) {
	x_cgo_ohos_environ = environ;
}
void (*x_cgo_inittls)(void **tlsg, void **tlsbase) = inittls;
