// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

#include <stdint.h>
#include <stdlib.h>

void* hitsumabushi_sysReserveOS(void* v, uintptr_t n);

void* hitsumabushi_sysAllocOS(uintptr_t n) {
  return hitsumabushi_sysReserveOS(NULL, n);
}

void hitsumabushi_sysUnusedOS(void* v, uintptr_t n) {
}

void hitsumabushi_sysUsedOS(void* v, uintptr_t n) {
}

void hitsumabushi_sysHugePageOS(void* v, uintptr_t n) {
}

void hitsumabushi_sysFreeOS(void* v, uintptr_t n) {
}

void hitsumabushi_sysFaultOS(void* v, uintptr_t n) {
}

void* hitsumabushi_sysReserveOS(void* v, uintptr_t n) {
  if (v) {
    return NULL;
  }
  return calloc(n, 1);
}

void hitsumabushi_sysMapOS(void* v, uintptr_t n) {
}
