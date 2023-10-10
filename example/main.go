package main

import (
	"fmt"

	. "github.com/spearson78/icgo"
)

//compile with no inlining and no optimizations
//go build -gcflags "-N -l"  .

func myfunc2(a int64, b int64) int64 {
	return a + b
}

func myfunc(param1 int64, param2 int64, param3 int64) int64 {
	local1 := param3 + param1
	return IC(myfunc2(myfunc2(param2, param3), local1))
}

func main() {
	fmt.Printf("Result: %v\n", myfunc(1, 2, 3))
}
