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

package binary_info

import (
	"sort"
	"strings"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)

type constantsMap map[dwarfRef]*constantType

type constantType struct {
	initialized bool
	values      []constantValue
}

type constantValue struct {
	name      string
	fullName  string
	value     int64
	singleBit bool
}

type constantValuesByValue []constantValue

func (v constantValuesByValue) Len() int               { return len(v) }
func (v constantValuesByValue) Less(i int, j int) bool { return v[i].value < v[j].value }
func (v constantValuesByValue) Swap(i int, j int)      { v[i], v[j] = v[j], v[i] }

func (cm constantsMap) Get(typ godwarf.Type) *constantType {
	ctyp := cm[dwarfRef{typ.Common().Index, typ.Common().Offset}]
	if ctyp == nil {
		return nil
	}
	typepkg := packageName(typ.String()) + "."
	if !ctyp.initialized {
		ctyp.initialized = true
		sort.Sort(constantValuesByValue(ctyp.values))
		for i := range ctyp.values {
			if strings.HasPrefix(ctyp.values[i].name, typepkg) {
				ctyp.values[i].name = ctyp.values[i].name[len(typepkg):]
			}
			if Popcnt(uint64(ctyp.values[i].value)) == 1 {
				ctyp.values[i].singleBit = true
			}
		}
	}
	return ctyp
}




func Popcnt(x uint64) int {
	const m0 = 0x5555555555555555 
	const m1 = 0x3333333333333333 
	const m2 = 0x0f0f0f0f0f0f0f0f 
	const m = 1<<64 - 1
	x = x>>1&(m0&m) + x&(m0&m)
	x = x>>2&(m1&m) + x&(m1&m)
	x = (x>>4 + x) & (m2 & m)
	x += x >> 8
	x += x >> 16
	x += x >> 32
	return int(x) & (1<<7 - 1)
}

func packageName(name string) string {
	pathend := strings.LastIndex(name, "/")
	if pathend < 0 {
		pathend = 0
	}

	if i := strings.Index(name[pathend:], "."); i != -1 {
		return name[:pathend+i]
	}
	return ""
}

func (ctyp *constantType) Describe(n int64) string {
	for _, val := range ctyp.values {
		if val.value == n {
			return val.name
		}
	}

	if n == 0 {
		return ""
	}

	
	

	fields := []string{}
	for _, val := range ctyp.values {
		if !val.singleBit {
			continue
		}
		if n&val.value != 0 {
			fields = append(fields, val.name)
			n = n & ^val.value
		}
	}
	if n == 0 {
		return strings.Join(fields, "|")
	}
	return ""
}
