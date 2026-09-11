package main

/*
#include <stdlib.h>
int callback_thread(int);
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

//export GoCallback
func GoCallback(x C.int) C.int {
	runtime.GC()
	return x * 3
}

//export GoCheckEnvironment
func GoCheckEnvironment() C.int {
	if os.Getenv("OHOS_GO_INTEROP") == "inherited" {
		return 1
	}
	return 0
}

//export GoCheckArgs
func GoCheckArgs(argc C.int, argv **C.char) C.int {
	if int(argc) != len(os.Args) {
		return 0
	}
	for i, arg := range unsafe.Slice(argv, int(argc)) {
		if os.Args[i] != C.GoString(arg) {
			return 0
		}
	}
	return 1
}

func main() {
	for i := 0; i < 100; i++ {
		if got := C.callback_thread(C.int(i)); got != C.int(i*3) {
			panic(fmt.Sprintf("C thread callback: got %d want %d", got, i*3))
		}
	}
	fmt.Printf("PASS cgo C-thread callbacks and GC (%s/%s, calls=%d)\n", runtime.GOOS, runtime.GOARCH, runtime.NumCgoCall())
}
