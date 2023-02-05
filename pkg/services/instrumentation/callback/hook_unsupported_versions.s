//go:build (!amd64 && !arm64) || !go1.15 || go1.20
// +build !amd64,!arm64 !go1.15 go1.20

#include "funcdata.h"
#include "textflag.h"

TEXT ·ShouldRunPrologue(SB), $0
RET
