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
