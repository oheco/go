package main

import (
	"bytes"
	"example.org/harmony-smoke/constraints"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	caseName = flag.String("case", "all", "case to run: all, hello, goroutines, gc, file, timer, tcp")
	repeat   = flag.Int("repeat", 1, "number of times to run the selected case(s)")
	timeout  = flag.Duration("timeout", 30*time.Second, "overall process timeout")
)

type testCase struct {
	name string
	run  func() error
}

func main() {
	flag.Parse()
	if *repeat < 1 || *repeat > 1000 {
		fail(fmt.Errorf("-repeat must be between 1 and 1000"))
	}
	if *timeout <= 0 {
		fail(fmt.Errorf("-timeout must be positive"))
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(*timeout):
			fmt.Fprintf(os.Stderr, "FAIL watchdog: exceeded %s\n", *timeout)
			os.Exit(2)
		case <-done:
		}
	}()

	tests := []testCase{
		{"hello", testHello},
		{"goroutines", testGoroutines},
		{"gc", testGC},
		{"file", testFile},
		{"timer", testTimer},
		{"tcp", testTCP},
	}
	selected := tests
	if *caseName != "all" {
		selected = nil
		for _, tc := range tests {
			if tc.name == *caseName {
				selected = append(selected, tc)
			}
		}
		if len(selected) == 0 {
			fail(fmt.Errorf("unknown -case %q", *caseName))
		}
	}

	for i := 1; i <= *repeat; i++ {
		for _, tc := range selected {
			if err := tc.run(); err != nil {
				fail(fmt.Errorf("%s (iteration %d): %w", tc.name, i, err))
			}
			fmt.Printf("PASS %s iteration=%d\n", tc.name, i)
		}
	}
	close(done)
	fmt.Println("PASS smoke")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "FAIL", err)
	os.Exit(1)
}

func testHello() error {
	fmt.Printf("Go=%s GOOS=%s GOARCH=%s CPUs=%d\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	if runtime.GOOS != "ohos" || runtime.GOARCH != "arm64" {
		return fmt.Errorf("want ohos/arm64, got %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if !constraints.Ohos || !constraints.LinuxABI || !constraints.Unix {
		return fmt.Errorf("incorrect platform constraints")
	}
	return nil
}

func testGoroutines() error {
	const workers, perWorker = 64, 1000
	values := make(chan int, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			sum := 0
			for n := 1; n <= perWorker; n++ {
				sum += base*perWorker + n
			}
			values <- sum
		}(worker)
	}
	go func() { wg.Wait(); close(values) }()
	got := 0
	for value := range values {
		got += value
	}
	total := workers * perWorker
	want := total * (total + 1) / 2
	if got != want {
		return fmt.Errorf("sum=%d, want %d", got, want)
	}
	return nil
}

func testGC() error {
	const blocks, size = 256, 16 * 1024
	kept := make([][]byte, blocks)
	for i := range kept {
		kept[i] = make([]byte, size)
		for j := range kept[i] {
			kept[i][j] = byte(i + j)
		}
	}
	runtime.GC()
	for i := range kept {
		for j := 0; j < size; j += 257 {
			if got, want := kept[i][j], byte(i+j); got != want {
				return fmt.Errorf("corruption at block=%d offset=%d: got %d, want %d", i, j, got, want)
			}
		}
	}
	runtime.KeepAlive(kept)
	return nil
}

func testFile() error {
	dir, err := os.MkdirTemp("", "go-ohos-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "data")
	want := []byte("HarmonyOS Go file smoke\n")
	if err := os.WriteFile(path, want, 0600); err != nil {
		return err
	}
	got, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("read data mismatch")
	}
	return nil
}

func testTimer() error {
	start := time.Now()
	timer := time.NewTimer(20 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		if time.Since(start) < 10*time.Millisecond {
			return fmt.Errorf("timer fired too early")
		}
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timer did not fire")
	}
}

func testTCP() error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()

	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		message := make([]byte, 4)
		if _, err := io.ReadFull(conn, message); err != nil {
			serverErr <- err
			return
		}
		if !bytes.Equal(message, []byte("ping")) {
			serverErr <- fmt.Errorf("server received %q", message)
			return
		}
		_, err = conn.Write([]byte("pong"))
		serverErr <- err
	}()

	conn, err := net.DialTimeout("tcp4", listener.Addr().String(), 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write([]byte("ping")); err != nil {
		return err
	}
	reply := make([]byte, 4)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if !bytes.Equal(reply, []byte("pong")) {
		return fmt.Errorf("client received %q", reply)
	}
	if err := <-serverErr; err != nil {
		return err
	}
	return nil
}
