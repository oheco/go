package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRoot(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := root.Mkdir("child", 0700); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile("child/data", []byte("ohos"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := root.ReadFile("child/data"); err != nil || string(got) != "ohos" {
		t.Fatalf("ReadFile = %q, %v", got, err)
	}
	if err := os.Symlink("child", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := root.Mkdir("link/", 0700); !os.IsExist(err) {
		t.Fatalf("Mkdir of symlink with trailing slash = %v, want EEXIST", err)
	}
}

func TestSignal(t *testing.T) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	defer signal.Stop(ch)
	if err := syscall.Kill(os.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("SIGUSR1 was not delivered")
	}
}

func TestSendFile(t *testing.T) {
	want := bytes.Repeat([]byte("ohos sendfile\n"), 8192)
	path := filepath.Join(t.TempDir(), "data")
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	sent := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(3 * time.Second))
			_, err = io.Copy(conn, file)
		}
		sent <- err
	}()
	conn, err := net.DialTimeout("tcp4", listener.Addr().String(), 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	got, err := io.ReadAll(conn)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("file transfer: received %d bytes, error %v", len(got), err)
	}
	if err := <-sent; err != nil {
		t.Fatal(err)
	}
}
