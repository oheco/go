package main

import (
	"fmt"
	"os"
	"plugin"
)

func main() {
	p, err := plugin.Open(os.Args[1])
	if err != nil { panic(err) }
	s, err := p.Lookup("Answer")
	if err != nil { panic(err) }
	if got := s.(func(int) int)(21); got != 42 { panic(got) }
	p2, err := plugin.Open(os.Args[1])
	if err != nil || p2 != p { panic("plugin cache mismatch") }
	fmt.Println("PASS plugin load, exported function and repeated load")
}
