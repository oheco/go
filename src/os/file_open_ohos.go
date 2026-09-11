// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/poll"
	"syscall"
)

func open(path string, flag int, perm uint32) (int, poll.SysFile, error) {
	// OHOS can create a regular file for "missing/", even with O_DIRECTORY.
	// Resolve terminal links ourselves when creating, so a slash in a link's
	// target is not lost either. O_NOFOLLOW prevents truncating or creating
	// through a link before its target has been checked.
	for links := 0; ; links++ {
		name := path
		flags := flag
		if flag&O_CREATE != 0 {
			if len(name) > 0 && IsPathSeparator(name[len(name)-1]) {
				name += "." // Require an existing directory in the kernel lookup.
			} else {
				flags |= syscall.O_NOFOLLOW
			}
		}
		fd, err := syscall.Open(name, flags, perm)
		if err != syscall.ELOOP || flag&O_CREATE == 0 || flag&(syscall.O_NOFOLLOW|O_EXCL) != 0 {
			return fd, poll.SysFile{}, err
		}
		if links >= 40 {
			return -1, poll.SysFile{}, syscall.ELOOP
		}
		target, err := Readlink(path)
		if err != nil {
			return -1, poll.SysFile{}, underlyingError(err)
		}
		path = joinSymlinkPath(path, target)
	}
}
