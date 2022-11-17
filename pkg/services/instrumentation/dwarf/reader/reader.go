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
