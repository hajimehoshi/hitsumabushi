// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

// This file defines C functions and system calls for Cgo.

#include <stddef.h> // for size_t
#include <signal.h> // for sigset_t and struct sigaction

#include "libcgo.h"
#include "libcgo_unix.h"

typedef unsigned int gid_t;

void *mmap(void *addr, size_t length, int prot, int flags, int fd, off_t offset) {
  abort();
  return NULL;
}

int munmap(void *addr, size_t length) {
  abort();
  return 0;
}

int pthread_sigmask(int how, const sigset_t *set, sigset_t *oldset) {
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

int sigaction(int signum, const struct sigaction *act, struct sigaction *oldact) {
  // Do nothing.
  return 0;
}

int sigaddset(sigset_t *set, int signum) {
  // Do nothing.
  return 0;
}

int sigemptyset(sigset_t *set) {
  // Do nothing.
  return 0;
}

int sigfillset(sigset_t *set) {
  // Do nothing.
  return 0;
}

int sigismember(const sigset_t *set, int signum) {
  // Do nothing.
  return 0;
}
