//
// https://go.dev/doc/asm
//
#include "textflag.h"
#include "go_tls.h"

TEXT ·getG(SB), NOSPLIT, $0-4
    MOVW g, R8
    MOVW R8, ret+0(FP)
    RET
