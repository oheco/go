// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package work

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"cmd/go/internal/cfg"
)

// Detect the actual host separately: both the Linux compatibility bootstrap
// and the ohos toolchain can run on either Linux or HarmonyOS.
var harmonyHost = sync.OnceValue(func() bool {
	if (runtime.GOOS != "linux" && runtime.GOOS != "ohos") || runtime.GOARCH != "arm64" {
		return false
	}
	out, err := exec.Command("uname", "-s").Output()
	if err != nil {
		return false
	}
	switch strings.TrimSpace(string(out)) {
	case "HarmonyOS", "OHOS", "OpenHarmony":
		return true
	}
	return false
})

// signHarmonyBinary signs a finished executable without changing its name.
// The original is preserved if signing fails. The caller must not modify the
// executable after this function returns, including rewriting its build ID.
func signHarmonyBinary(target string) error {
	tool, err := exec.LookPath("binary-sign-tool")
	if err != nil {
		return fmt.Errorf("HarmonyOS signing requires binary-sign-tool in PATH; check the LLVM tool directory: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(target), ".go-sign-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	signed := filepath.Join(dir, "signed")
	out, err := exec.Command(tool, "sign", "-inFile", target, "-outFile", signed, "-selfSign", "1").CombinedOutput()
	if err != nil {
		return fmt.Errorf("signing %s: %w\n%s", target, err, out)
	}
	if err := os.Chmod(signed, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(signed, target)
}

// Keep signed host executables distinct from unsigned Linux cache entries.
func harmonySigningRequired() bool {
	if (cfg.Goos != "linux" && cfg.Goos != "ohos") || cfg.Goarch != "arm64" {
		return false
	}
	switch cfg.BuildBuildmode {
	case "default", "exe", "pie", "c-shared", "shared", "plugin":
		return harmonyHost()
	}
	return false
}

func (b *Builder) signHarmonyAction(a *Action, target string) error {
	if (a.Mode != "link" && a.Mode != "go build -buildmode=shared") || !harmonySigningRequired() {
		return nil
	}
	if cfg.BuildX || cfg.BuildN {
		b.Shell(a).ShowCmd("", "binary-sign-tool sign -inFile %s -outFile <temporary> -selfSign 1 # atomic replacement", target)
	}
	if cfg.BuildN {
		return nil
	}
	return signHarmonyBinary(target)
}
