#  IcGo

**WARNING: This library is in an alpha proof of concept state and ONLY works in a very limited scope**

IceCream like debug logging for Go. Inspired by the python IceCream [https://github.com/gruns/icecream]

# Why

Even though go has an excellent debugger (delve) I still often find myself inserting logging statements into code. These logs usually just report the variables and return of a line of code that seems to be failing. After watching a video about Python's IceCream library I thought it would be great to have something similar in Go.

I found this project linked from the Python IceCream github repo. [https://github.com/WAY29/icecream-go] unfortunately this was missing a critical feature. The replacement of parameters and variables in the output with their current values.

Knowing that a debugger has access to this information I thought that it must be possible to do this at runtime in Go so I set about trying to make it happen. I was able to produce a very limited proof of concept.

# Example

**Currently only tested on linux amd64***

```go
package main

import (
	"fmt"
	. "github.com/spearson78/icgo"
)

func myfunc2(a int64, b int64) int64 {
	return a + b
}

func myfunc(param1 int64, param2 int64, param3 int64) int64 {
	local1 := param3 + param1
	return IC(myfunc2(myfunc2(param2, param3), local1)) //This line will be logged
}

func main() {
	fmt.Printf("Result: %v\n", myfunc(1, 2, 3))
}
```

Compiling this with ```go build -gcflags "-N -l" -o bin/example ./example/.```

And running the binary will produce the following output

```2023/10/13 09:41:24 ic|         return 9 <- IC(myfunc2(myfunc2(2, 1), 4))```

Here the parameters and local variables have been replaced with their runtime values and the return is indicated with the ```<-``` symbol

# How

When logging the Runtime.caller is used to get the calling functions filename,line number and program counter.

The DWARF debugging data is then loaded. The program counter is used to get the functions debugging data.

This data describes the 

1. the name and stack location of all local variables 
2. the location and names of all parameters

Parameters are forced to spill onto the stack and their locations are inferred from the Go internal ABI spec [https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md]

the corresponding line is then read from the source file and the parameter values and return value are replaced and the line is logged.

# Current Limitations

The current state of the impementation is a proof of concept to gauge whether there is any interest in this functionality.

The current limitations are:
- Linux amd64 only
- Only int64 parameters and variables are supported
- Single return value only
- Extremely inefficient
- Dependent on the Go internal ABI spec
- Bare minimum support of location opcodes
- Primitive parameter replacement in the logged string
- Inlining and optimizations must be disabled during compilation

With some additional effort most of these limitations can be overcome. 
