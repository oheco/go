// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package work

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignHarmonyBinary(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("requires a POSIX shell")
	}
	for _, test := range []string{"success", "failure", "missing"} {
		t.Run(test, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "go")
			if err := os.WriteFile(target, []byte("original"), 0755); err != nil {
				t.Fatal(err)
			}
			tool := filepath.Join(dir, "binary-sign-tool")
			if test != "missing" {
				script := "#!" + sh + "\nwhile [ \"$#\" -gt 0 ]; do\nif [ \"$1\" = -outFile ]; then shift; output=$1; fi\nshift\ndone\necho signed > \"$output\"\n"
				if test == "failure" {
					script += "echo deliberate-failure >&2\nexit 17\n"
				}
				if err := os.WriteFile(tool, []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", dir)
			err := signHarmonyBinary(target)
			want := "original"
			if test == "success" {
				want = "signed\n"
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("expected signing error")
			} else if test == "failure" && !strings.Contains(err.Error(), "deliberate-failure") {
				t.Fatalf("missing signer diagnostic: %v", err)
			}
			got, err := os.ReadFile(target)
			if err != nil || string(got) != want {
				t.Fatalf("target = %q, %v; want %q", got, err, want)
			}
			info, err := os.Stat(target)
			if err != nil || info.Mode().Perm() != 0755 {
				t.Fatalf("executable permissions changed: %v, %v", info, err)
			}
			left, err := filepath.Glob(filepath.Join(dir, ".go-sign-*"))
			if err != nil || len(left) != 0 {
				t.Fatalf("temporary signing files remain: %v, %v", left, err)
			}
		})
	}
}
