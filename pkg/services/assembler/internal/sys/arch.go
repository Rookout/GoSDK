// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package sys

import "encoding/binary"



type ArchFamily byte

const (
	NoArch ArchFamily = iota
	AMD64
	ARM
	ARM64
	I386
	Loong64
	MIPS
	MIPS64
	PPC64
	RISCV64
	S390X
	Wasm
)


type Arch struct {
	Name   string
	Family ArchFamily

	ByteOrder binary.ByteOrder

	
	
	PtrSize int

	
	RegSize int

	
	MinLC int

	
	
	
	
	Alignment int8

	
	
	
	CanMergeLoads bool

	
	
	CanJumpTable bool

	
	
	HasLR bool

	
	
	
	
	
	
	FixedFrameSize int64
}



func (a *Arch) InFamily(xs ...ArchFamily) bool {
	for _, x := range xs {
		if a.Family == x {
			return true
		}
	}
	return false
}

var Arch386 = &Arch{
	Name:           "386",
	Family:         I386,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        4,
	RegSize:        4,
	MinLC:          1,
	Alignment:      1,
	CanMergeLoads:  true,
	HasLR:          false,
	FixedFrameSize: 0,
}

var ArchAMD64 = &Arch{
	Name:           "amd64",
	Family:         AMD64,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          1,
	Alignment:      1,
	CanMergeLoads:  true,
	CanJumpTable:   true,
	HasLR:          false,
	FixedFrameSize: 0,
}

var ArchARM = &Arch{
	Name:           "arm",
	Family:         ARM,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        4,
	RegSize:        4,
	MinLC:          4,
	Alignment:      4, 
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 4, 
}

var ArchARM64 = &Arch{
	Name:           "arm64",
	Family:         ARM64,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          4,
	Alignment:      1,
	CanMergeLoads:  true,
	CanJumpTable:   true,
	HasLR:          true,
	FixedFrameSize: 8, 
}

var ArchLoong64 = &Arch{
	Name:           "loong64",
	Family:         Loong64,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          4,
	Alignment:      8, 
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 8, 
}

var ArchMIPS = &Arch{
	Name:           "mips",
	Family:         MIPS,
	ByteOrder:      binary.BigEndian,
	PtrSize:        4,
	RegSize:        4,
	MinLC:          4,
	Alignment:      4,
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 4, 
}

var ArchMIPSLE = &Arch{
	Name:           "mipsle",
	Family:         MIPS,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        4,
	RegSize:        4,
	MinLC:          4,
	Alignment:      4,
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 4, 
}

var ArchMIPS64 = &Arch{
	Name:           "mips64",
	Family:         MIPS64,
	ByteOrder:      binary.BigEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          4,
	Alignment:      8,
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 8, 
}

var ArchMIPS64LE = &Arch{
	Name:           "mips64le",
	Family:         MIPS64,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          4,
	Alignment:      8,
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 8, 
}

var ArchPPC64 = &Arch{
	Name:          "ppc64",
	Family:        PPC64,
	ByteOrder:     binary.BigEndian,
	PtrSize:       8,
	RegSize:       8,
	MinLC:         4,
	Alignment:     1,
	CanMergeLoads: false,
	HasLR:         true,
	
	
	FixedFrameSize: 4 * 8,
}

var ArchPPC64LE = &Arch{
	Name:           "ppc64le",
	Family:         PPC64,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          4,
	Alignment:      1,
	CanMergeLoads:  true,
	HasLR:          true,
	FixedFrameSize: 4 * 8,
}

var ArchRISCV64 = &Arch{
	Name:           "riscv64",
	Family:         RISCV64,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          4,
	Alignment:      8, 
	CanMergeLoads:  false,
	HasLR:          true,
	FixedFrameSize: 8, 
}

var ArchS390X = &Arch{
	Name:           "s390x",
	Family:         S390X,
	ByteOrder:      binary.BigEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          2,
	Alignment:      1,
	CanMergeLoads:  true,
	HasLR:          true,
	FixedFrameSize: 8, 
}

var ArchWasm = &Arch{
	Name:           "wasm",
	Family:         Wasm,
	ByteOrder:      binary.LittleEndian,
	PtrSize:        8,
	RegSize:        8,
	MinLC:          1,
	Alignment:      1,
	CanMergeLoads:  false,
	HasLR:          false,
	FixedFrameSize: 0,
}

var Archs = [...]*Arch{
	Arch386,
	ArchAMD64,
	ArchARM,
	ArchARM64,
	ArchLoong64,
	ArchMIPS,
	ArchMIPSLE,
	ArchMIPS64,
	ArchMIPS64LE,
	ArchPPC64,
	ArchPPC64LE,
	ArchRISCV64,
	ArchS390X,
	ArchWasm,
}
