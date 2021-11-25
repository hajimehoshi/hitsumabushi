//--from
TEXT runtime·open(SB),NOSPLIT|NOFRAME,$0-20
	MOVD	$AT_FDCWD, R0
	MOVD	name+0(FP), R1
	MOVW	mode+8(FP), R2
	MOVW	perm+12(FP), R3
	MOVD	$SYS_openat, R8
	SVC
	CMN	$4095, R0
	BCC	done
	MOVW	$-1, R0
done:
	MOVW	R0, ret+16(FP)
	RET
//--to
TEXT runtime·open_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	12(R0), R2
	MOVD	0(R0), R0
	BL	c_open(SB)
	RET
//--from
TEXT runtime·closefd(SB),NOSPLIT|NOFRAME,$0-12
	MOVW	fd+0(FP), R0
	MOVD	$SYS_close, R8
	SVC
	CMN	$4095, R0
	BCC	done
	MOVW	$-1, R0
done:
	MOVW	R0, ret+8(FP)
	RET
//--to
TEXT runtime·closefd_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	c_closefd(SB)
	RET
//--from
TEXT runtime·write1(SB),NOSPLIT|NOFRAME,$0-28
	MOVD	fd+0(FP), R0
	MOVD	p+8(FP), R1
	MOVW	n+16(FP), R2
	MOVD	$SYS_write, R8
	SVC
	MOVW	R0, ret+24(FP)
	RET
//--to
TEXT runtime·write1_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	16(R0), R2
	MOVD	0(R0), R0
	BL	c_write1(SB)
	RET
//--from
TEXT runtime·read(SB),NOSPLIT|NOFRAME,$0-28
	MOVW	fd+0(FP), R0
	MOVD	p+8(FP), R1
	MOVW	n+16(FP), R2
	MOVD	$SYS_read, R8
	SVC
	MOVW	R0, ret+24(FP)
	RET
//--to
TEXT runtime·read_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	16(R0), R2
	MOVD	0(R0), R0
	BL	c_read(SB)
	RET
//--from
TEXT runtime·usleep(SB),NOSPLIT,$24-4
	MOVWU	usec+0(FP), R3
	MOVD	R3, R5
	MOVW	$1000000, R4
	UDIV	R4, R3
	MOVD	R3, 8(RSP)
	MUL	R3, R4
	SUB	R4, R5
	MOVW	$1000, R4
	MUL	R4, R5
	MOVD	R5, 16(RSP)

	// nanosleep(&ts, 0)
	ADD	$8, RSP, R0
	MOVD	$0, R1
	MOVD	$SYS_nanosleep, R8
	SVC
	RET
//--to
TEXT runtime·usleep_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	c_usleep(SB)
	RET
//--from
TEXT runtime·gettid(SB),NOSPLIT,$0-4
	MOVD	$SYS_gettid, R8
	SVC
	MOVW	R0, ret+0(FP)
	RET
//--to
TEXT runtime·gettid_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	c_gettid(SB)
	RET
//--from
TEXT runtime·nanotime1(SB),NOSPLIT,$24-8
	MOVD	RSP, R20	// R20 is unchanged by C code
	MOVD	RSP, R1

	MOVD	g_m(g), R21	// R21 = m

	// Set vdsoPC and vdsoSP for SIGPROF traceback.
	// Save the old values on stack and restore them on exit,
	// so this function is reentrant.
	MOVD	m_vdsoPC(R21), R2
	MOVD	m_vdsoSP(R21), R3
	MOVD	R2, 8(RSP)
	MOVD	R3, 16(RSP)

	MOVD	LR, m_vdsoPC(R21)
	MOVD	R20, m_vdsoSP(R21)

	MOVD	m_curg(R21), R0
	CMP	g, R0
	BNE	noswitch

	MOVD	m_g0(R21), R3
	MOVD	(g_sched+gobuf_sp)(R3), R1	// Set RSP to g0 stack

noswitch:
	SUB	$32, R1
	BIC	$15, R1
	MOVD	R1, RSP

	MOVW	$CLOCK_MONOTONIC, R0
	MOVD	runtime·vdsoClockgettimeSym(SB), R2
	CBZ	R2, fallback

	// Store g on gsignal's stack, so if we receive a signal
	// during VDSO code we can find the g.
	// If we don't have a signal stack, we won't receive signal,
	// so don't bother saving g.
	// When using cgo, we already saved g on TLS, also don't save
	// g here.
	// Also don't save g if we are already on the signal stack.
	// We won't get a nested signal.
	MOVBU	runtime·iscgo(SB), R22
	CBNZ	R22, nosaveg
	MOVD	m_gsignal(R21), R22          // g.m.gsignal
	CBZ	R22, nosaveg
	CMP	g, R22
	BEQ	nosaveg
	MOVD	(g_stack+stack_lo)(R22), R22 // g.m.gsignal.stack.lo
	MOVD	g, (R22)

	BL	(R2)

	MOVD	ZR, (R22)  // clear g slot, R22 is unchanged by C code

	B	finish

nosaveg:
	BL	(R2)
	B	finish

fallback:
	MOVD	$SYS_clock_gettime, R8
	SVC

finish:
	MOVD	0(RSP), R3	// sec
	MOVD	8(RSP), R5	// nsec

	MOVD	R20, RSP	// restore SP
	// Restore vdsoPC, vdsoSP
	// We don't worry about being signaled between the two stores.
	// If we are not in a signal handler, we'll restore vdsoSP to 0,
	// and no one will care about vdsoPC. If we are in a signal handler,
	// we cannot receive another signal.
	MOVD	16(RSP), R1
	MOVD	R1, m_vdsoSP(R21)
	MOVD	8(RSP), R1
	MOVD	R1, m_vdsoPC(R21)

	// sec is in R3, nsec in R5
	// return nsec in R3
	MOVD	$1000000000, R4
	MUL	R4, R3
	ADD	R5, R3
	MOVD	R3, ret+0(FP)
	RET
//--to
//--from
TEXT runtime·sigaltstack(SB),NOSPLIT|NOFRAME,$0
	MOVD	new+0(FP), R0
	MOVD	old+8(FP), R1
	MOVD	$SYS_sigaltstack, R8
	SVC
	CMN	$4095, R0
	BCC	ok
	MOVD	$0, R0
	MOVD	R0, (R0)	// crash
ok:
	RET
//--to
//--from
TEXT runtime·osyield(SB),NOSPLIT|NOFRAME,$0
	MOVD	$SYS_sched_yield, R8
	SVC
	RET
//--to
//--from
TEXT runtime·sched_getaffinity(SB),NOSPLIT|NOFRAME,$0
	MOVD	pid+0(FP), R0
	MOVD	len+8(FP), R1
	MOVD	buf+16(FP), R2
	MOVD	$SYS_sched_getaffinity, R8
	SVC
	MOVW	R0, ret+24(FP)
	RET
//--to
//--append
TEXT runtime·malloc_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	0(R0), R0
	BL	c_malloc(SB)
	RET

TEXT runtime·nanotime1_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	c_nanotime1(SB)
	RET

TEXT runtime·sched_getaffinity_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	16(R0), R2
	MOVD	0(R0), R0
	BL	c_sched_getaffinity(SB)
	RET

TEXT runtime·osyield_trampoline(SB),NOSPLIT,$0
	BL	c_osyield(SB)
	RET

TEXT runtime·pthread_mutex_init_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	0(R0), R0
	BL	pthread_mutex_init(SB)
	RET

TEXT runtime·pthread_mutex_lock_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	pthread_mutex_lock(SB)
	RET

TEXT runtime·pthread_mutex_unlock_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	pthread_mutex_unlock(SB)
	RET

TEXT runtime·pthread_cond_init_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	0(R0), R0
	BL	pthread_cond_init(SB)
	RET

TEXT runtime·pthread_cond_wait_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	0(R0), R0
	BL	pthread_cond_wait(SB)
	RET

TEXT runtime·pthread_cond_signal_trampoline(SB),NOSPLIT,$0
	MOVD	0(R0), R0
	BL	pthread_cond_signal(SB)
	RET

TEXT runtime·pthread_cond_timedwait_relative_np_trampoline(SB),NOSPLIT,$0
	MOVD	8(R0), R1
	MOVD	16(R0), R2
	MOVD	0(R0), R0
	BL	c_pthread_cond_timedwait_relative_np(SB)
	RET
