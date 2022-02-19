//
// https://go.dev/doc/asm
//
#include "textflag.h"
#include "go_tls.h"

TEXT ·getG(SB), NOSPLIT, $0-4
    get_tls(AX)
    MOVQ    AX, ret+0(FP)
    RET
