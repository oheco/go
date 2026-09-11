// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build cgo

package runtime_test

import "testing"

func TestCgoOhosConcurrentFork(t *testing.T) {
	got := runTestProg(t, "testprogcgo", "OhosConcurrentFork")
	if got != "OK\n" {
		t.Fatalf("concurrent fork corrupted signal handling:\n%s", got)
	}
}
