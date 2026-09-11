// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

//go:linkname _cgo_ohos_load_g _cgo_ohos_load_g
var _cgo_ohos_load_g unsafe.Pointer

//go:linkname _cgo_ohos_save_g _cgo_ohos_save_g
var _cgo_ohos_save_g unsafe.Pointer

//go:linkname _cgo_ohos_environ _cgo_ohos_environ
var _cgo_ohos_environ unsafe.Pointer

func goenvsOhosLibrary() {
	goargsOhosLibrary()
	if _cgo_ohos_environ == nil {
		return
	}
	ep := *(***byte)(_cgo_ohos_environ)
	if ep == nil {
		return
	}
	n := int32(0)
	for argv_index(ep, n) != nil {
		n++
	}
	envs = make([]string, n)
	for i := int32(0); i < n; i++ {
		envs[i] = gostring(argv_index(ep, i))
	}
}

// musl does not supply argc/argv to library constructors. Recover a snapshot
// of the host's argument vector before package initialization. As with auxv,
// a host without accessible procfs leaves the vector empty.
func goargsOhosLibrary() {
	path := []byte("/proc/self/cmdline\x00")
	fd := open(&path[0], 0, 0)
	if fd < 0 {
		return
	}
	defer closefd(fd)
	var data []byte
	var buf [4096]byte
	for {
		n := read(fd, noescape(unsafe.Pointer(&buf[0])), int32(len(buf)))
		if n == -_EINTR {
			continue
		}
		if n < 0 {
			return
		}
		if n == 0 {
			break
		}
		data = append(data, buf[:n]...)
	}
	start := 0
	for i, b := range data {
		if b == 0 {
			argslice = append(argslice, string(data[start:i]))
			start = i + 1
		}
	}
}
