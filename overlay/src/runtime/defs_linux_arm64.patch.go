//--append
// The definition of these constants and the sizes of these structs depends on environments.

const _ETIMEDOUT = 110

type pthread uintptr
type pthreadattr struct {
	X_opaque [56]uint8
}
type pthreadmutex struct {
	X_opaque [32]uint8
}
type pthreadmutexattr struct {
	X_opaque [4]uint8
}
type pthreadcond struct {
	X_opaque [32]uint8
}
type pthreadcondattr struct {
	X_opaque [4]uint8
}
