//go:build arm64 && go1.15 && !go1.21
// +build arm64,go1.15,!go1.21

#include "funcdata.h"
#include "textflag.h"



TEXT ·ShouldRunPrologue(SB), NOFRAME|NOSPLIT, $0
NO_LOCAL_POINTERS
ADD $900, R19 
SUB R19, RSP, R19 
MOVD 0x10(g), R20 
CMP R20, R19 
BGE NoRunPrologue 
RunPrologue:
MOVD $0x1, R19
RET
NoRunPrologue:
MOVD $0x0, R19
RET


TEXT ·getContext(SB), NOSPLIT, $0-8
MOVD R19, ret+0(FP)
RET

