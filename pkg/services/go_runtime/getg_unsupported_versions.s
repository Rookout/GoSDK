//go:build !go1.15 || go1.20 || !amd64
// +build !go1.15 go1.20 !amd64

#include "funcdata.h"
#include "textflag.h"

TEXT ·Getg(SB),NOSPLIT, $0-8
RET
