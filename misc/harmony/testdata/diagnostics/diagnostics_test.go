package diagnostics

import (
	"bytes"
	"os"
	"runtime/pprof"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	got, err := RoundTrip([]byte("ohos"))
	if err != nil || string(got) != "ohos" { t.Fatalf("%q, %v", got, err) }
}

func TestHeapProfile(t *testing.T) {
	for _, debug := range []int{0, 1, 2} {
		var buf bytes.Buffer
		if err := pprof.Lookup("heap").WriteTo(&buf, debug); err != nil { t.Fatal(err) }
		if buf.Len() == 0 { t.Fatal("empty heap profile") }
	}
	if path := os.Getenv("OHOS_HEAP_PROFILE"); path != "" {
		f, err := os.Create(path)
		if err != nil { t.Fatal(err) }
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil { t.Fatal(err) }
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	data := bytes.Repeat([]byte("ohos"), 1024)
	for b.Loop() { _, _ = RoundTrip(data) }
}

func FuzzRoundTrip(f *testing.F) {
	f.Add([]byte("ohos"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := RoundTrip(data)
		if err != nil || !bytes.Equal(got, data) { t.Fatalf("round trip: %v", err) }
	})
}
