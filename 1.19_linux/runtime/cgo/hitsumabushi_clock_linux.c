// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

#include <stdint.h>
#include <time.h> // for clock_gettime
#include <unistd.h> // for usleep

void hitsumabushi_usleep(useconds_t usec) {
  usleep(usec);
}

void hitsumabushi_walltime1(int64_t* sec, int32_t* nsec) {
  struct timespec tp;
  clock_gettime(CLOCK_REALTIME, &tp);
  *sec = tp.tv_sec;
  *nsec = tp.tv_nsec;
}

int64_t hitsumabushi_nanotime1() {
  struct timespec tp;
  clock_gettime(CLOCK_MONOTONIC, &tp);
  return (int64_t)(tp.tv_sec) * 1000000000ll + (int64_t)tp.tv_nsec;
}
