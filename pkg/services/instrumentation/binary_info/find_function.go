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
	"sort"
	"strings"

	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/line"
)



func (bi *BinaryInfo) FindFileLocation(filename string, lineno int) ([]uint64, string, *Function, rookoutErrors.RookoutError) {
	possibleCUs, matchedFile, searchFileErr := bi.getBestMatchingFile(filename)
	if searchFileErr != nil {
		return nil, "", nil, searchFileErr
	}
	// got a file
	filename = matchedFile
	
	
	

	
	var pcs []line.PCStmt
	for _, cu := range possibleCUs {
		if cu.lineInfo == nil || cu.lineInfo.Lookup[filename] == nil {
			continue
		}

		pcs = append(pcs, cu.lineInfo.LineToPCs(filename, lineno)...)
	}

	if len(pcs) == 0 {
		
		
		
		
		for _, pc := range bi.inlinedCallLines[fileLine{filename, lineno}] {
			pcs = append(pcs, line.PCStmt{PC: pc, Stmt: true})
		}
	}

	if len(pcs) == 0 {
		return nil, "", nil, rookoutErrors.NewLineNotFound(filename, lineno)
	}

	

	pcByFunc := map[*Function][]line.PCStmt{}
	sort.Slice(pcs, func(i, j int) bool { return pcs[i].PC < pcs[j].PC })
	var fn *Function
	for _, pcstmt := range pcs {
		if fn == nil || (pcstmt.PC < fn.Entry) || (pcstmt.PC >= fn.End) {
			fn = bi.PCToFunc(pcstmt.PC)
		}
		if fn != nil {
			pcByFunc[fn] = append(pcByFunc[fn], pcstmt)
		}
	}

	var selectedPCs []uint64

	for fn, pcs := range pcByFunc {

		

		if strings.Contains(fn.Name, "·dwrap·") || fn.trampoline {
			
			continue
		}

		dwtree, err := fn.cu.image.GetDwarfTree(fn.Offset)
		if err != nil {
			return nil, "", nil, rookoutErrors.NewFailedToGetDWARFTree(err)
		}
		inlrngs := allInlineCallRanges(dwtree)

		
		
		
		
		findInlRng := func(pc uint64) dwarf.Offset {
			for _, inlrng := range inlrngs {
				if inlrng.rng[0] <= pc && pc < inlrng.rng[1] {
					return inlrng.off
				}
			}
			return fn.Offset
		}

		pcsByOff := map[dwarf.Offset][]line.PCStmt{}

		for _, pc := range pcs {
			off := findInlRng(pc.PC)
			pcsByOff[off] = append(pcsByOff[off], pc)
		}

		
		
		

		for off, pcs := range pcsByOff {
			sort.Slice(pcs, func(i, j int) bool { return pcs[i].PC < pcs[j].PC })

			var selectedPC uint64
			for _, pc := range pcs {
				if pc.Stmt {
					selectedPC = pc.PC
					break
				}
			}

			if selectedPC == 0 && len(pcs) > 0 {
				selectedPC = pcs[0].PC
			}

			if selectedPC == 0 {
				continue
			}

			

			if off == fn.Offset && fn.Entry == selectedPC {
				return bi.FindFileLocation(filename, lineno+1)
			}

			selectedPCs = append(selectedPCs, selectedPC)
		}
	}

	sort.Slice(selectedPCs, func(i, j int) bool { return selectedPCs[i] < selectedPCs[j] })

	return selectedPCs, filename, fn, nil
}


type inlRange struct {
	off   dwarf.Offset
	depth uint32
	rng   [2]uint64
}





func allInlineCallRanges(tree *godwarf.Tree) []inlRange {
	var r []inlRange

	var visit func(*godwarf.Tree, uint32)
	visit = func(n *godwarf.Tree, depth uint32) {
		if n.Tag == dwarf.TagInlinedSubroutine {
			for _, rng := range n.Ranges {
				r = append(r, inlRange{off: n.Offset, depth: depth, rng: rng})
			}
		}
		for _, child := range n.Children {
			visit(child, depth+1)
		}
	}
	visit(tree, 0)

	sort.SliceStable(r, func(i, j int) bool { return r[i].depth > r[j].depth })
	return r
}
