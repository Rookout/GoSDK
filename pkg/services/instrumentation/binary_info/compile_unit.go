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
	"debug/dwarf"

	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/line"
)

type compileUnit struct {
	name    string 
	Version uint8  
	lowPC   uint64
	ranges  [][2]uint64

	entry     *dwarf.Entry        
	IsGo      bool                
	lineInfo  *line.DebugLineInfo 
	optimized bool                
	producer  string              

	offset dwarf.Offset 

	image *Image 
}

type compileUnitsByOffset []*compileUnit

func (v compileUnitsByOffset) Len() int               { return len(v) }
func (v compileUnitsByOffset) Less(i int, j int) bool { return v[i].offset < v[j].offset }
func (v compileUnitsByOffset) Swap(i int, j int)      { v[i], v[j] = v[j], v[i] }

func (c *compileUnit) pcInRange(pc uint64) bool {
	for _, rng := range c.ranges {
		if pc >= rng[0] && pc < rng[1] {
			return true
		}
	}
	return false
}
