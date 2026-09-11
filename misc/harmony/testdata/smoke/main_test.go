package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestRuntime(t *testing.T) {
	for _, tc := range []testCase{
		{"hello", testHello},
		{"goroutines", testGoroutines},
		{"gc", testGC},
		{"file", testFile},
		{"timer", testTimer},
		{"tcp", testTCP},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSubprocess(t *testing.T) {
	if os.Getenv("HARMONY_SMOKE_CHILD") == "1" {
		os.Exit(23)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestSubprocess$")
	cmd.Env = append(os.Environ(), "HARMONY_SMOKE_CHILD=1")
	out, err := cmd.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 23 {
		t.Fatalf("child exit: %v, output: %s", err, out)
	}
}
