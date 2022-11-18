// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

#include <stddef.h> // for size_t
#include <stdint.h> // for int32_t
#include <sched.h> // for pid_t

int32_t hitsumabushi_getproccount() {
	return 1;
}

int32_t hitsumabushi_sched_getaffinity(pid_t pid, size_t cpusetsize, void *mask) {
    int32_t numcpu = hitsumabushi_getproccount();
    for (int32_t i = 0; i < numcpu; i += 8)
        ((unsigned char*)mask)[i / 8] = (unsigned char)((1u << (numcpu - i)) - 1);
    // https://man7.org/linux/man-pages/man2/sched_setaffinity.2.html
    // > On success, the raw sched_getaffinity() system call returns the
    // > number of bytes placed copied into the mask buffer;
    return (numcpu + 7) / 8;
}
