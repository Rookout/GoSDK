//go:build go1.16 && !go1.23 && amd64
// +build go1.16,!go1.23,amd64

#include "funcdata.h"
#include "textflag.h"

































TEXT ·Getg(SB),NOSPLIT, $0-8
MOVQ TLS, CX
MOVQ 0(CX)(TLS*1), AX
MOVQ AX, ret+0(FP)
RET
