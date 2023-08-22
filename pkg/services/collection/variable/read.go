// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package variable

import (
	"encoding/binary"
	"fmt"

	"github.com/Rookout/GoSDK/pkg/config"
	"github.com/Rookout/GoSDK/pkg/services/collection/memory"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)



func readIntRaw(mem memory.MemoryReader, addr uint64, size int64) (int64, error) {
	var n int64

	val := make([]byte, int(size))
	_, err := mem.ReadMemory(val, addr)
	if err != nil {
		return 0, err
	}

	switch size {
	case 1:
		n = int64(int8(val[0]))
	case 2:
		n = int64(int16(binary.LittleEndian.Uint16(val)))
	case 4:
		n = int64(int32(binary.LittleEndian.Uint32(val)))
	case 8:
		n = int64(binary.LittleEndian.Uint64(val))
	}

	return n, nil
}

func readUintRaw(mem memory.MemoryReader, addr uint64, size int64) (uint64, error) {
	var n uint64

	val := make([]byte, int(size))
	_, err := mem.ReadMemory(val, addr)
	if err != nil {
		return 0, err
	}

	switch size {
	case 1:
		n = uint64(val[0])
	case 2:
		n = uint64(binary.LittleEndian.Uint16(val))
	case 4:
		n = uint64(binary.LittleEndian.Uint32(val))
	case 8:
		n = uint64(binary.LittleEndian.Uint64(val))
	}

	return n, nil
}

func readStringInfo(mem memory.MemoryReader, bi *binary_info.BinaryInfo, addr uint64, typ *godwarf.StringType) (uint64, int64, error) {
	
	

	mem = memory.CacheMemory(mem, addr, bi.PointerSize*2)

	var strlen int64
	var outaddr uint64
	var err error

	for _, field := range typ.StructType.Field {
		switch field.Name {
		case "len":
			strlen, err = readIntRaw(mem, addr+uint64(field.ByteOffset), int64(bi.PointerSize))
			if err != nil {
				return 0, 0, fmt.Errorf("could not read string len %s", err)
			}
			if strlen < 0 {
				return 0, 0, fmt.Errorf("invalid length: %d", strlen)
			}
		case "str":
			outaddr, err = readUintRaw(mem, addr+uint64(field.ByteOffset), int64(bi.PointerSize))
			if err != nil {
				return 0, 0, fmt.Errorf("could not read string pointer %s", err)
			}
			if addr == 0 {
				return 0, 0, nil
			}
		}
	}

	return outaddr, strlen, nil
}

func readStringValue(mem memory.MemoryReader, addr uint64, strlen int64, cfg config.ObjectDumpConfig) (string, error) {
	if strlen == 0 {
		return "", nil
	}

	count := strlen
	if count > int64(cfg.MaxString) {
		count = int64(cfg.MaxString)
	}

	val := make([]byte, int(count))
	_, err := mem.ReadMemory(val, addr)
	if err != nil {
		return "", fmt.Errorf("could not read string at %#v due to %s", addr, err)
	}

	return string(val), nil
}

func readCStringValue(mem memory.MemoryReader, addr uint64, cfg config.ObjectDumpConfig) (string, bool, error) {
	buf := make([]byte, cfg.MaxString) 
	val := buf[:0]                     

	for len(buf) > 0 {
		
		
		
		
		
		
		
		curaddr := addr + uint64(len(val))
		maxsize := int(alignAddr(int64(curaddr+1), 1024) - int64(curaddr))
		size := len(buf)
		if size > maxsize {
			size = maxsize
		}

		_, err := mem.ReadMemory(buf[:size], curaddr)
		if err != nil {
			return "", false, fmt.Errorf("could not read string at %#v due to %s", addr, err)
		}

		done := false
		for i := 0; i < size; i++ {
			if buf[i] == 0 {
				done = true
				size = i
				break
			}
		}

		val = val[:len(val)+size]
		buf = buf[size:]
		if done {
			return string(val), true, nil
		}
	}

	return string(val), false, nil
}
