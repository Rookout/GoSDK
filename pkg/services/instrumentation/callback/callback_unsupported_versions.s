//go:build (!amd64 && !arm64) || !go1.15 || go1.21
// +build !amd64,!arm64 !go1.15 go1.21

#include "funcdata.h"
#include "textflag.h"

TEXT ·getContext(SB), NOSPLIT, $0-8
RET
