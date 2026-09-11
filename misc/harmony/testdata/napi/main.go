package main

/*
#cgo LDFLAGS: -lace_napi.z
#include <stdint.h>
*/
import "C"

import "runtime"

//export GoAdd
func GoAdd(a, b C.int32_t) C.int64_t {
	return C.int64_t(int64(a) + int64(b))
}

//export GoSum
func GoSum(n C.int32_t) C.int64_t {
	// Exercise Go allocation and GC from an Ark native worker thread.
	values := make([]int64, int(n))
	var total int64
	for i := range values {
		values[i] = int64(i + 1)
		total += values[i]
	}
	runtime.GC()
	runtime.KeepAlive(values)
	return C.int64_t(total)
}

func main() {}
