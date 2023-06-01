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

package reader

import (
	"debug/dwarf"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
)




func InlineStack(root *godwarf.Tree, pc uint64) []*godwarf.Tree {
	v := []*godwarf.Tree{}
	for _, child := range root.Children {
		v = inlineStackInternal(v, child, pc)
	}
	return v
}














func inlineStackInternal(stack []*godwarf.Tree, n *godwarf.Tree, pc uint64) []*godwarf.Tree {
	switch n.Tag {
	case dwarf.TagSubprogram, dwarf.TagInlinedSubroutine, dwarf.TagLexDwarfBlock:
		if pc == 0 || n.ContainsPC(pc) {
			for _, child := range n.Children {
				stack = inlineStackInternal(stack, child, pc)
			}
			if n.Tag == dwarf.TagInlinedSubroutine {
				stack = append(stack, n)
			}
		}
	}
	return stack
}
