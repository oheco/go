package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	dir, err := os.MkdirTemp("", "go-mmap-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	for _, mode := range []os.FileMode{0600, 0755} {
		name := fmt.Sprintf("%s/%o", dir, mode)
		f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			panic(err)
		}
		fmt.Printf("mode=%o fallocate=%v truncate=%v\n", mode, syscall.Fallocate(int(f.Fd()), 0, 0, 8192), f.Truncate(8192))
		for _, flags := range []int{syscall.MAP_SHARED, syscall.MAP_PRIVATE} {
			buf, err := syscall.Mmap(int(f.Fd()), 0, 8192, syscall.PROT_READ|syscall.PROT_WRITE, flags)
			fmt.Printf("mmap flags=%d err=%v\n", flags, err)
			if err == nil {
				buf[0] = 42
				fmt.Printf("munmap=%v\n", syscall.Munmap(buf))
			}
		}
		f.Close()
	}
}
