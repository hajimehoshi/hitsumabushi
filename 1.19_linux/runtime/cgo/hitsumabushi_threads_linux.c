// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

#include <errno.h>
#include <pthread.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static const int kPseudoFutexWait = 0;
static const int kPseudoFutexWake = 1;

static void pseudo_futex(uint32_t *uaddr, int mode, uint32_t val, const struct timespec *reltime) {
  static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
  static pthread_cond_t cond = PTHREAD_COND_INITIALIZER;

  struct timespec abstime;
  if (reltime) {
    // We are not sure CLOCK_REALTIME is correct or not here.
    // However, this time limit is actually not used as the condition variable is shared by
    // all the threads. Before the time limit reaches, the thread wakes up in 99.9999...% cases.
    clock_gettime(CLOCK_REALTIME, &abstime);
    abstime.tv_sec += reltime->tv_sec;
    abstime.tv_nsec += reltime->tv_nsec;
    if (1000000000 <= abstime.tv_nsec) {
      abstime.tv_sec += 1;
      abstime.tv_nsec -= 1000000000;
    }
  }

  int ret = pthread_mutex_lock(&mutex);
  if (ret) {
    fprintf(stderr, "pthread_mutex_lock failed: %d\n", ret);
    abort();
  }

  switch (mode) {
  case kPseudoFutexWait:
    if (reltime) {
      uint32_t v = 0;
      __atomic_load(uaddr, &v, __ATOMIC_RELAXED);
      if (v == val) {
        int ret = pthread_cond_timedwait(&cond, &mutex, &abstime);
        if (ret && ret != ETIMEDOUT) {
          fprintf(stderr, "pthread_cond_timedwait failed: %d\n", ret);
          abort();
        }
      }
    } else {
      uint32_t v = 0;
      __atomic_load(uaddr, &v, __ATOMIC_RELAXED);
      if (v == val) {
        int ret = pthread_cond_wait(&cond, &mutex);
        if (ret) {
          fprintf(stderr, "pthread_cond_wait failed: %d\n", ret);
          abort();
        }
      }
    }
    break;
  case kPseudoFutexWake:
    if (val != 1) {
      fprintf(stderr, "val for waking must be 1 but %d\n", val);
      abort();
    }
    // TODO: broadcasting is not efficient. Use a mutex for each uaddr.
    int ret = pthread_cond_broadcast(&cond);
    if (ret) {
      fprintf(stderr, "pthread_cond_broadcast failed: %d\n", ret);
      abort();
    }
    break;
  }

  ret = pthread_mutex_unlock(&mutex);
  if (ret) {
    fprintf(stderr, "pthread_mutex_unlock failed: %d\n", ret);
    abort();
  }
}

int32_t hitsumabushi_futex(uint32_t *uaddr, int32_t futex_op, uint32_t val,
                           const struct timespec *timeout,
                           uint32_t *uaddr2, uint32_t val3) {
  enum {
    kFutexWaitPrivate = 128,
    kFutexWakePrivate = 129,
  };

  switch (futex_op) {
  case kFutexWaitPrivate:
    pseudo_futex(uaddr, kPseudoFutexWait, val, timeout);
    break;
  case kFutexWakePrivate:
    pseudo_futex(uaddr, kPseudoFutexWake, val, NULL);
    break;
  }

  // This function should return the number of awaken threads, but now it is impossible.
  // Just return 0.
  return 0;
}

uint32_t hitsumabushi_gettid() {
  uint64_t tid64 = (uint64_t)(pthread_self());
  uint32_t tid = (uint32_t)(tid64 >> 32) ^ (uint32_t)(tid64);
  return tid;
}

int32_t hitsumabushi_osyield() {
  return sched_yield();
}

void hitsumabushi_exit(int32_t code) {
  exit(code);
}
