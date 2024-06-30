//go:build amd64 && go1.16 && !go1.23
// +build amd64,go1.16,!go1.23

#include "funcdata.h"
#include "textflag.h"


TEXT ·getContext(SB), NOSPLIT, $0-8
MOVQ R12, ret+0(FP)
RET
