package icgo

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"ekyu.moe/leb128"
)

//go:noinline
func regSpill(p1 int64, p2 int64, p3 int64, p4 int64, p5 int64, p6 int64, p7 int64, p8 int64, p9 int64, p10 int64) int64 {
	return p1 + p2 + p3 + p4 + p4 + p5 + p6 + p7 + p8 + p9 + p10
}

func printStack(fp uintptr) {
	for i := -64; i < 64; i += 1 {
		p := fp + uintptr(i*8)
		val := *(*int)(unsafe.Pointer(p))
		log.Printf("STACK: %v %v : %v\n", i, p, val)
	}
}

//go:noinline
func IC[T any](v T) T {
	icfp := uintptr(unsafe.Pointer(&v))

	regSpill(100, 200, 300, 400, 500, 600, 700, 800, 900, 1000)

	fp := *(*uintptr)(unsafe.Pointer(icfp - 24))

	pc, file, lineNum, ok := runtime.Caller(1)
	if ok {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			log.Printf("ic| error reading source file: %v : %v", file, err)
			return v
		}

		lines := strings.Split(string(fileContent), "\n")

		line := lines[lineNum-1]
		vstr := fmt.Sprintf("%v <- IC(", v)
		vline := strings.Replace(line, "IC(", vstr, 1)

		elfFile, err := elf.Open(os.Args[0])
		if err != nil {
			log.Printf("ic| error reading executable file: %v : %v", os.Args[0], err)
			return v
		}

		dwarfData, err := elfFile.DWARF()
		if err != nil {
			log.Printf("ic| error reading parsing DWARF data: %v", err)
			return v
		}

		entryReader := dwarfData.Reader()

		for {
			entry, err := entryReader.Next()
			if entry == nil {
				break
			}
			if err == nil && entry.Tag == dwarf.TagSubprogram {
				nameAttr := entry.AttrField(dwarf.AttrName)
				lowPc := entry.AttrField(dwarf.AttrLowpc)
				highPc := entry.AttrField(dwarf.AttrHighpc)
				frameBase := entry.AttrField(dwarf.AttrFrameBase)
				if nameAttr != nil && lowPc != nil && highPc != nil && frameBase != nil {
					if uint64(pc) >= lowPc.Val.(uint64) && uint64(pc) <= highPc.Val.(uint64) {
						var paramNames []string
						if entry.Children {
							entry, err = entryReader.Next()
							for err == nil && entry.Tag != 0 {
								if entry.Tag == dwarf.TagFormalParameter {
									paramNames = append(paramNames, entry.AttrField(dwarf.AttrName).Val.(string))
								} else if entry.Tag == dwarf.TagVariable {
									name := entry.AttrField(dwarf.AttrName).Val.(string)
									location, _ := entry.AttrField(dwarf.AttrLocation).Val.([]uint8)

									if location[0] == 0x91 { //DW_OP_fbreg
										d, n := leb128.DecodeSleb128(location[1:])
										if n == 0 {
											log.Printf("ic| error decoding variable location: %v : %v", location[1:])
											return v
										}

										val := *(*int64)(unsafe.Pointer(fp + uintptr(d) + 16))
										vline = strings.ReplaceAll(vline, name, fmt.Sprintf("%v", val))
									} else {
										log.Printf("ic| unknown variable location: %02x : %v", location[0], err)
										return v
									}
								}
								entry, err = entryReader.Next()
							}
						}

						paramOffset := uintptr(8 * len(paramNames))
						for _, name := range paramNames {
							val := *(*int64)(unsafe.Pointer(fp + paramOffset))
							vline = strings.ReplaceAll(vline, name, fmt.Sprintf("%v", val))
							paramOffset = paramOffset - 8
						}

					}
				}
			}
		}

		log.Printf("ic| %v", vline)
	}

	return v
}
