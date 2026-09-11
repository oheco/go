// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// This program failed with SIGSEGV when run under the C/C++ ThreadSanitizer.
// The Go runtime had re-registered the C handler with the wrong flags due to a
// typo, resulting in null pointers being passed for the info and context
// parameters to the handler.

/*
#cgo CFLAGS: -fsanitize=thread
#cgo LDFLAGS: -fsanitize=thread

#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ucontext.h>

void check_params(int signo, siginfo_t *info, void *context) {
	if (info == NULL || context == NULL) {
		fprintf(stderr, "signal handler did not receive siginfo/ucontext.\n");
		abort();
	}
	ucontext_t* uc = (ucontext_t*)(context);

	if (info->si_signo != signo) {
		fprintf(stderr, "info->si_signo does not match signo.\n");
		abort();
	}

#if defined(__OHOS__) && defined(__aarch64__)
	// OHOS supplies a zero uc_stack even for a native C handler running
	// on a registered alternate stack. Check the saved PC instead, keeping
	// the assertion that Go forwarded a valid SA_SIGINFO context.
	if (uc->uc_mcontext.pc == 0) {
		fprintf(stderr, "ucontext has no saved PC.\n");
		abort();
	}
#else
	if (uc->uc_stack.ss_size == 0) {
		fprintf(stderr, "uc_stack has size 0.\n");
		abort();
	}
#endif
}


// Set up the signal handler in a high priority constructor, so
// that it is installed before the Go code starts.

static void register_handler(void) __attribute__ ((constructor (200)));

static void register_handler() {
	struct sigaction sa;
	memset(&sa, 0, sizeof(sa));
	sigemptyset(&sa.sa_mask);
	sa.sa_flags = SA_SIGINFO;
	sa.sa_sigaction = check_params;

	if (sigaction(SIGUSR1, &sa, NULL) != 0) {
		perror("failed to register SIGUSR1 handler");
		exit(EXIT_FAILURE);
	}
}
*/
import "C"

import "syscall"

func init() {
	C.raise(C.int(syscall.SIGUSR1))
}

func main() {}
