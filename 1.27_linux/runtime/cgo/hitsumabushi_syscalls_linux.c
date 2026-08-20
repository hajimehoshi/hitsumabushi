// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

// This file defines C functions and system calls for Cgo.

#include <pthread.h>
#include <errno.h>
#include <string.h>
#include <stdlib.h>
#include <stdatomic.h>
#include <unistd.h> // for usleep
#include <stddef.h> // for size_t

#include "libcgo.h"
#include "libcgo_unix.h"

typedef unsigned int gid_t;

extern int hitsumabushi_clock_gettime(clockid_t clk_id, struct timespec *tp);
extern int32_t hitsumabushi_getproccount();

void *mmap(void *addr, size_t length, int prot, int flags, int fd, off_t offset) {
  abort();
  return NULL;
}

int munmap(void *addr, size_t length) {
  abort();
  return 0;
}

int pthread_sigmask(int how, void *set, void *oldset) {
  // Do nothing.
  return 0;
}

int setegid(gid_t gid) {
  // Do nothing.
  return 0;
}

int seteuid(uid_t gid) {
  // Do nothing.
  return 0;
}

int setgid(gid_t gid) {
  // Do nothing.
  return 0;
}

int setgroups(size_t size, const gid_t *list) {
  // Do nothing.
  return 0;
}

int setregid(gid_t rgid, gid_t egid) {
  // Do nothing.
  return 0;
}

int setreuid(uid_t ruid, uid_t euid) {
  // Do nothing.
  return 0;
}

int setresgid(gid_t rgid, gid_t egid, gid_t sgid) {
  // Do nothing.
  return 0;
}

int setresuid(uid_t ruid, uid_t euid, uid_t suid) {
  // Do nothing.
  return 0;
}

int setuid(uid_t gid) {
  // Do nothing.
  return 0;
}

int sigaction(int signum, void *act, void *oldact) {
  // Do nothing.
  return 0;
}

int sigaddset(void *set, int signum) {
  // Do nothing.
  return 0;
}

int sigemptyset(void *set) {
  // Do nothing.
  return 0;
}

int sigfillset(void *set) {
  // Do nothing.
  return 0;
}

int sigismember(void *set, int signum) {
  // Do nothing.
  return 0;
}

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
