//go:build go1.16 && !go1.23 && arm64
// +build go1.16,!go1.23,arm64

#include "funcdata.h"
#include "textflag.h"




TEXT ·Getg(SB), NOSPLIT ,$-8-8
MOVD    g, ret+0(FP)
RET
