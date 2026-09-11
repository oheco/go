// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"internal/platform"
	"os"
	"path/filepath"
	"testing"
)

func TestOhosBuildConstraints(t *testing.T) {
	oldOS, oldArch := goos, goarch
	goos, goarch = "ohos", "arm64"
	t.Cleanup(func() { goos, goarch = oldOS, oldArch })
	dir := t.TempDir()
	for _, tt := range []struct {
		name, constraint string
		want             bool
	}{
		{"port_ohos.go", "ohos && unix", true},
		{"port_linux_arm64.go", "linux", true},
		{"port_linux.go", "linux && !ohos", false},
		{"port_android.go", "", false},
		{"port_ohos_amd64.go", "", false},
	} {
		file := filepath.Join(dir, tt.name)
		data := "package runtime\n"
		if tt.constraint != "" {
			data = "//go:build " + tt.constraint + "\n\n" + data
		}
		if err := os.WriteFile(file, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if got := shouldbuild(file, "runtime"); got != tt.want {
			t.Errorf("shouldbuild(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestMustLinkExternal verifies that the mustLinkExternal helper
// function matches internal/platform.MustLinkExternal.
func TestMustLinkExternal(t *testing.T) {
	for _, goos := range okgoos {
		for _, goarch := range okgoarch {
			for _, cgoEnabled := range []bool{true, false} {
				got := mustLinkExternal(goos, goarch, cgoEnabled)
				want := platform.MustLinkExternal(goos, goarch, cgoEnabled)
				if got != want {
					t.Errorf("mustLinkExternal(%q, %q, %v) = %v; want %v", goos, goarch, cgoEnabled, got, want)
				}
			}
		}
	}
}

func TestRequiredBootstrapVersion(t *testing.T) {
	testCases := map[string]string{
		"1.22": "1.20",
		"1.23": "1.20",
		"1.24": "1.22",
		"1.25": "1.22",
		"1.26": "1.24",
		"1.27": "1.24",
	}

	for v, want := range testCases {
		if got := requiredBootstrapVersion(v); got != want {
			t.Errorf("requiredBootstrapVersion(%v): got %v, want %v", v, got, want)
		}
	}
}
