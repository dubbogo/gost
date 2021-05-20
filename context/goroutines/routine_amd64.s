// +build amd64, gc

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"



// func goRoutine() *g
TEXT ·goRoutine(SB), NOSPLIT, $0-8
	MOVQ (TLS), AX
	MOVQ AX, ret+0(FP)
	RET

// func goRoutineID() int64
//TEXT ·goRoutineID(SB),NOSPLIT,$0-8
//	MOVQ (TLS), AX
//	MOVQ g(AX), AX
//	MOVQ g_goid(g), ret+0(FP)
//	RET
