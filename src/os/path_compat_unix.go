// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix || (js && wasm) || wasip1

package os

import "syscall"

// joinSymlinkPath replaces the final component with a symlink target.
// Do not clean the path: resolving a/../b requires traversing a, which may
// itself be a symlink. Preserve any trailing slash in target as well.
func joinSymlinkPath(name, target string) string {
	if len(target) > 0 && IsPathSeparator(target[0]) {
		return target
	}
	i := len(name) - 1
	for i >= 0 && !IsPathSeparator(name[i]) {
		i--
	}
	return name[:i+1] + target
}

// resolveOhosTrailingSlash gives Remove and Rename the same terminal-slash
// semantics as Root. The OHOS unlinkat/renameat syscalls otherwise discard
// the slash, potentially removing or renaming a non-directory instead.
// Like doInRoot, this resolves a terminal symlink only when a slash is present.
func resolveOhosTrailingSlash(name string, allowMissing bool) (string, error) {
	if len(name) == 0 || !IsPathSeparator(name[len(name)-1]) {
		return name, nil
	}
	for links := 0; ; links++ {
		for len(name) > 1 && IsPathSeparator(name[len(name)-1]) {
			name = name[:len(name)-1]
		}
		fi, err := Lstat(name)
		if err != nil {
			if allowMissing && IsNotExist(err) {
				return name, nil
			}
			return "", underlyingError(err)
		}
		if fi.IsDir() {
			return name, nil
		}
		if fi.Mode()&ModeSymlink == 0 {
			return "", syscall.ENOTDIR
		}
		if links >= 40 {
			return "", syscall.ELOOP
		}
		target, err := Readlink(name)
		if err != nil {
			return "", underlyingError(err)
		}
		name = joinSymlinkPath(name, target)
	}
}
