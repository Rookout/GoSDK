//go:build !go1.15 || go1.20 || !amd64
// +build !go1.15 go1.20 !amd64

#include "funcdata.h"
#include "textflag.h"

TEXT ·PrepForCallback(SB), $0
RET

TEXT ·MoreStack(SB), $0
RET

TEXT ·ShouldRunPrologue(SB), NOSPLIT, $0
RET
