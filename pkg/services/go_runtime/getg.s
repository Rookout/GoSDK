//go:build go1.15 && !go1.20 && amd64
// +build go1.15,!go1.20,amd64

#include "funcdata.h"
#include "textflag.h"

































TEXT ·Getg(SB),NOSPLIT, $0-8
MOVQ TLS, CX
MOVQ 0(CX)(TLS*1), AX
MOVQ AX, ret+0(FP)
RET
