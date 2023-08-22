//go:build arm64 && go1.15 && !go1.22
// +build arm64,go1.15,!go1.22

#include "funcdata.h"
#include "textflag.h"


TEXT ·getContext(SB), NOSPLIT, $0-8
MOVD R19, ret+0(FP)
RET

