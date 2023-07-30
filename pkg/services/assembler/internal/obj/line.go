// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.assembler file.

package obj

import (
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/goobj"
	"github.com/Rookout/GoSDK/pkg/services/assembler/internal/src"
)


func (ctxt *Link) AddImport(pkg string, fingerprint goobj.FingerprintType) {
	ctxt.Imports = append(ctxt.Imports, goobj.ImportedPkg{Pkg: pkg, Fingerprint: fingerprint})
}




func (ctxt *Link) getFileSymbolAndLine(xpos src.XPos) (f string, l int32) {
	pos := ctxt.InnermostPos(xpos)
	if !pos.IsKnown() {
		pos = src.Pos{}
	}
	return pos.SymFilename(), int32(pos.RelLine())
}





func (ctxt *Link) getFileIndexAndLine(xpos src.XPos) (int, int32) {
	f, l := ctxt.getFileSymbolAndLine(xpos)
	return ctxt.PosTable.FileIndex(f), l
}
