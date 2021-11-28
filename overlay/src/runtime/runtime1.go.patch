//--from
func argv_index(argv **byte, i int32) *byte {
	return *(**byte)(add(unsafe.Pointer(argv), uintptr(i)*sys.PtrSize))
}
//--to
func argv_index(argv **byte, i int32) *byte {
	// Unfortunately, the given args are not reliable on some machines. Always return nil.
	return nil
}
//--from
func args(c int32, v **byte) {
	argc = c
	argv = v
	sysargs(c, v)
}
//--to
func args(c int32, v **byte) {
	// Unfortunately, the given args are not reliable on some machines. Ignore them.
	argc = 0
	argv = nil
	sysargs(c, v)
}
