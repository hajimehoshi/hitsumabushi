// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

// This file defines C functions and system calls for Cgo.

#include <pthread.h> // for pthread_self
#include <stdint.h> // for uint32_t etc.
#include <stdlib.h> // for exit()
#include <unistd.h> // for usleep
#include <stddef.h> // for size_t
#include <time.h> // for struct timespec

#include "libcgo.h"
#include "libcgo_unix.h"

extern int hitsumabushi_clock_gettime(clockid_t clk_id, struct timespec *tp);
extern int32_t hitsumabushi_getproccount();

uint32_t hitsumabushi_gettid() {
  uint64_t tid64 = (uint64_t)(pthread_self());
  uint32_t tid = (uint32_t)(tid64 >> 32) ^ (uint32_t)(tid64);
  return tid;
}

int64_t hitsumabushi_nanotime1() {
  struct timespec tp;
  hitsumabushi_clock_gettime(CLOCK_MONOTONIC, &tp);
  return (int64_t)(tp.tv_sec) * 1000000000ll + (int64_t)tp.tv_nsec;
}

int32_t hitsumabushi_osyield() {
  return sched_yield();
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

void hitsumabushi_usleep(useconds_t usec) {
  usleep(usec);
}

void hitsumabushi_walltime1(int64_t* sec, int32_t* nsec) {
  struct timespec tp;
  hitsumabushi_clock_gettime(CLOCK_REALTIME, &tp);
  *sec = tp.tv_sec;
  *nsec = tp.tv_nsec;
}

void hitsumabushi_exit(int32_t code) {
  exit(code);
}
