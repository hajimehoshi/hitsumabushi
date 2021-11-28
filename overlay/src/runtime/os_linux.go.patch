//--from
import (
	"runtime/internal/sys"
	"unsafe"
)
//--to
import (
	"internal/abi"
	"unsafe"
)
//--from
//go:noescape
func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) int32
//--to
//go:nosplit
//go:cgo_unsafe_args
func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) int32 {
	return libcCall(unsafe.Pointer(abi.FuncPCABI0(futex_trampoline)), unsafe.Pointer(&addr))
}
func futex_trampoline()
//--from
func sysargs(argc int32, argv **byte) {
	n := argc + 1

	// skip over argv, envp to get to auxv
	for argv_index(argv, n) != nil {
		n++
	}

	// skip NULL separator
	n++

	// now argv+n is auxv
	auxv := (*[1 << 28]uintptr)(add(unsafe.Pointer(argv), uintptr(n)*sys.PtrSize))
	if sysauxv(auxv[:]) != 0 {
		return
	}
	// In some situations we don't get a loader-provided
	// auxv, such as when loaded as a library on Android.
	// Fall back to /proc/self/auxv.
	fd := open(&procAuxv[0], 0 /* O_RDONLY */, 0)
	if fd < 0 {
		// On Android, /proc/self/auxv might be unreadable (issue 9229), so we fallback to
		// try using mincore to detect the physical page size.
		// mincore should return EINVAL when address is not a multiple of system page size.
		const size = 256 << 10 // size of memory region to allocate
		p, err := mmap(nil, size, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
		if err != 0 {
			return
		}
		var n uintptr
		for n = 4 << 10; n < size; n <<= 1 {
			err := mincore(unsafe.Pointer(uintptr(p)+n), 1, &addrspace_vec[0])
			if err == 0 {
				physPageSize = n
				break
			}
		}
		if physPageSize == 0 {
			physPageSize = size
		}
		munmap(p, size)
		return
	}
	var buf [128]uintptr
	n = read(fd, noescape(unsafe.Pointer(&buf[0])), int32(unsafe.Sizeof(buf)))
	closefd(fd)
	if n < 0 {
		return
	}
	// Make sure buf is terminated, even if we didn't read
	// the whole file.
	buf[len(buf)-2] = _AT_NULL
	sysauxv(buf[:])
}
//--to
func sysargs(argc int32, argv **byte) {
	// argc/argv is not reliable on some machines.
	// Skip analysing them.

	// In some situations we don't get a loader-provided
	// auxv, such as when loaded as a library on Android.
	// Fall back to /proc/self/auxv.
	fd := open(&procAuxv[0], 0 /* O_RDONLY */, 0)
	if fd < 0 {
		// On Android, /proc/self/auxv might be unreadable (issue 9229), so we fallback to
		// try using mincore to detect the physical page size.
		// mincore should return EINVAL when address is not a multiple of system page size.
		const size = 256 << 10 // size of memory region to allocate
		p, err := mmap(nil, size, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
		if err != 0 {
			return
		}
		var n uintptr
		for n = 4 << 10; n < size; n <<= 1 {
			err := mincore(unsafe.Pointer(uintptr(p)+n), 1, &addrspace_vec[0])
			if err == 0 {
				physPageSize = n
				break
			}
		}
		if physPageSize == 0 {
			physPageSize = size
		}
		munmap(p, size)
		return
	}
	var buf [128]uintptr
	n := read(fd, noescape(unsafe.Pointer(&buf[0])), int32(unsafe.Sizeof(buf)))
	closefd(fd)
	if n < 0 {
		return
	}
	// Make sure buf is terminated, even if we didn't read
	// the whole file.
	buf[len(buf)-2] = _AT_NULL
	sysauxv(buf[:])
}
//--from
func getRandomData(r []byte) {
	if startupRandomData != nil {
		n := copy(r, startupRandomData)
		extendRandom(r, n)
		return
	}
	fd := open(&urandom_dev[0], 0 /* O_RDONLY */, 0)
	n := read(fd, unsafe.Pointer(&r[0]), int32(len(r)))
	closefd(fd)
	extendRandom(r, int(n))
}
//--to
// Use getRandomData in os_plan9.go.

//go:nosplit
func getRandomData(r []byte) {
	// inspired by wyrand see hash32.go for detail
	t := nanotime()
	v := getg().m.procid ^ uint64(t)

	for len(r) > 0 {
		v ^= 0xa0761d6478bd642f
		v *= 0xe7037ed1a0b428db
		size := 8
		if len(r) < 8 {
			size = len(r)
		}
		for i := 0; i < size; i++ {
			r[i] = byte(v >> (8 * i))
		}
		r = r[size:]
		v = v>>32 | v<<32
	}
}
//--from
func gettid() uint32
//--to
//go:nosplit
//go:cgo_unsafe_args
func gettid() uint64 {
	var ret uint64
	var ret2 = &ret
	libcCall(unsafe.Pointer(abi.FuncPCABI0(gettid_trampoline)), unsafe.Pointer(&ret2))
	return ret
}
func gettid_trampoline();
//--from
//go:noescape
func sigaltstack(new, old *stackt)
//--to
func sigaltstack(new, old *stackt) {
	// Do nothing.
}
//--from
func sigprocmask(how int32, new, old *sigset) {
	rtsigprocmask(how, new, old, int32(unsafe.Sizeof(*new)))
}
//--to
func sigprocmask(how int32, new, old *sigset) {
	// Do nothing.
	// rtsigprocmask(how, new, old, int32(unsafe.Sizeof(*new)))
}
//--from
//go:noescape
func sched_getaffinity(pid, len uintptr, buf *byte) int32
//--to
//go:nosplit
//go:cgo_unsafe_args
func sched_getaffinity(pid, len uintptr, buf *byte) int32 {
	return libcCall(unsafe.Pointer(abi.FuncPCABI0(sched_getaffinity_trampoline)), unsafe.Pointer(&pid))
}
func sched_getaffinity_trampoline()
//--from
func osyield()
//--to
//go:nosplit
//go:cgo_unsafe_args
func osyield() {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(osyield_trampoline)), nil)
}
func osyield_trampoline()
