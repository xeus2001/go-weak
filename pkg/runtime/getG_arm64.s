//
// https://go.dev/doc/asm
//
#include "textflag.h"
#include "go_tls.h"

TEXT ·getG(SB), NOSPLIT, $0-8
    MOVD g, R8
    MOVD R8, ret+0(FP)
    RET
