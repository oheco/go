package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"unsafe"
)

func main() {
	mode := "safe"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "safe":
		p := C.calloc(1, 16)
		if p == nil { panic("allocation failed") }
		b := unsafe.Slice((*byte)(p), 16)
		b[0] = 42
		fmt.Println(b[0])
		C.free(p)
	case "race":
		var x int
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10000; j++ { x++ }
			}()
		}
		wg.Wait()
		fmt.Println(x)
	case "asan":
		p := C.malloc(8)
		if p == nil { panic("allocation failed") }
		b := unsafe.Slice((*byte)(p), 16)
		b[12] = 42 // Deliberate heap overflow; the detector must reject it.
		runtime.KeepAlive(b)
		C.free(p)
	case "msan":
		p := C.malloc(8)
		if p == nil { panic("allocation failed") }
		fmt.Println(*(*uint64)(p)) // Deliberate uninitialized read.
		C.free(p)
	default:
		panic("unknown detector test")
	}
}
