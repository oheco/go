// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

import "testing"

func test6997(t *testing.T) {
	t.Skip("OHOS libc does not provide pthread_cancel")
}
