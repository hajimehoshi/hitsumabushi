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
  // The returned pointer must be aligned to 1 << 9 bytes.
  // See also:
  // * https://cs.opensource.google/go/go/+/refs/tags/go1.25.0:src/runtime/tagptr_64bit.go
  // * https://go.dev/cl/665815/
  return aligned_alloc(1 << 9, n);
}

void hitsumabushi_sysMapOS(void* v, uintptr_t n) {
}
