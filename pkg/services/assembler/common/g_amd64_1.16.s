//go:build amd64
// +build amd64

#include "funcdata.h"
#include "textflag.h"

TEXT ·movGToR12(SB),NOSPLIT, $0
MOVQ (TLS), R12
RET
