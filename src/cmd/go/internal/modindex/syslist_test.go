// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is a lightly modified copy go/build/syslist_test.go.

package modindex

import (
	"go/build"
	"runtime"
	"testing"
)

var (
	thisOS    = runtime.GOOS
	thisArch  = runtime.GOARCH
	otherOS   = anotherOS()
	otherArch = anotherArch()
)

func anotherOS() string {
	if thisOS != "darwin" && thisOS != "ios" {
		return "darwin"
	}
	return "linux"
}

func anotherArch() string {
	if thisArch != "amd64" {
		return "amd64"
	}
	return "386"
}

type GoodFileTest struct {
	name   string
	result bool
}

var tests = []GoodFileTest{
	{"file.go", true},
	{"file.c", true},
	{"file_foo.go", true},
	{"file_" + thisArch + ".go", true},
	{"file_" + otherArch + ".go", false},
	{"file_" + thisOS + ".go", true},
	{"file_" + otherOS + ".go", false},
	{"file_" + thisOS + "_" + thisArch + ".go", true},
	{"file_" + otherOS + "_" + thisArch + ".go", false},
	{"file_" + thisOS + "_" + otherArch + ".go", false},
	{"file_" + otherOS + "_" + otherArch + ".go", false},
	{"file_foo_" + thisArch + ".go", true},
	{"file_foo_" + otherArch + ".go", false},
	{"file_" + thisOS + ".c", true},
	{"file_" + otherOS + ".c", false},
}

func TestGoodOSArch(t *testing.T) {
	for _, test := range tests {
		if (*Context)(&build.Default).goodOSArchFile(test.name, make(map[string]bool)) != test.result {
			t.Fatalf("goodOSArchFile(%q) != %v", test.name, test.result)
		}
	}
}

func TestOhosConstraints(t *testing.T) {
	ctxt := (*Context)(&build.Context{GOOS: "ohos", GOARCH: "arm64"})
	for _, tag := range []string{"ohos", "linux", "unix", "arm64"} {
		if !ctxt.matchTag(tag, nil) {
			t.Errorf("ohos/arm64 does not match %q", tag)
		}
	}
	for _, tt := range []GoodFileTest{
		{"file_ohos.go", true},
		{"file_ohos_arm64.go", true},
		{"file_ohos_arm64.syso", true},
		{"file_linux_arm64.syso", false},
		{"file_linux.syso", false},
		{"file_arm64.syso", true},
		{"file_linux_arm64.go", true},
		{"file_android.go", false},
		{"file_ohos_amd64.go", false},
	} {
		if got := ctxt.goodOSArchFile(tt.name, nil); got != tt.result {
			t.Errorf("goodOSArchFile(%q) = %v, want %v", tt.name, got, tt.result)
		}
	}
	linux := (*Context)(&build.Context{GOOS: "linux", GOARCH: "arm64"})
	if linux.goodOSArchFile("file_ohos.go", nil) || linux.matchTag("ohos", nil) {
		t.Fatal("Linux must not select ohos-only source")
	}
}
