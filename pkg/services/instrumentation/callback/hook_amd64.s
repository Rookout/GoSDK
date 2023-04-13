//go:build amd64 && go1.15 && !go1.21
// +build amd64,go1.15,!go1.21

#include "funcdata.h"
#include "textflag.h"




TEXT ·ShouldRunPrologue(SB), NOSPLIT, $0
NO_LOCAL_POINTERS
ADDQ $100, R12 
MOVQ SP, R13 
SUBQ R12, R13 
MOVQ (TLS), R12 
MOVQ 0x10(R12), R12 
CMPQ R13, R12
JGE NoRunPrologue 
RunPrologue:
MOVQ $0x1, R13
RET
NoRunPrologue:
MOVQ $0x0, R13
RET


TEXT ·getContext(SB), NOSPLIT, $0-8
MOVQ R12, ret+0(FP)
RET
