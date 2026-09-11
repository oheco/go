// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
#include <signal.h>
#include <stddef.h>
#include <stdint.h>

static uintptr_t segv_handler(void) {
	struct sigaction sa;
	if (sigaction(SIGSEGV, NULL, &sa) != 0)
		return (uintptr_t)-1;
	return (uintptr_t)sa.sa_sigaction;
}
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

func init() {
	register("OhosConcurrentFork", ohosConcurrentFork)
	register("OhosForkChild", func() {})
}

func ohosConcurrentFork() {
	self, err := os.Executable()
	if err != nil {
		panic(err)
	}
	want := C.segv_handler()
	if want == 0 || want == ^C.uintptr_t(0) {
		panic("missing SIGSEGV handler")
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 16 {
		wg.Go(func() {
			<-start
			for range 100 {
				if err := exec.Command(self, "OhosForkChild").Run(); err != nil {
					panic(err)
				}
				if got := C.segv_handler(); got != want {
					panic(fmt.Sprintf("SIGSEGV handler changed after fork: %#x -> %#x", want, got))
				}
			}
		})
	}
	close(start)
	wg.Wait()

	// Verify that the preserved handler still converts a Go fault to panic.
	defer func() {
		if recover() == nil {
			panic("nil fault did not panic")
		}
		fmt.Println("OK")
	}()
	var p *int
	*p = 1
}
