// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Hitsumabushi Authors

// This file defines C functions and system calls for Cgo.

#include <pthread.h>
#include <errno.h>
#include <string.h>
#include <stdlib.h>
#include <stdatomic.h>
#include <fcntl.h>
#include <sys/stat.h>

#include "libcgo.h"
#include "libcgo_unix.h"

static const int kFDOffset = 100;

typedef struct {
  const void* content;
  size_t      content_size;
  size_t      current;
  int32_t     fd;
} pseudo_file;

// TODO: Do we need to protect this by mutex?
static pseudo_file pseudo_files[100];

static pthread_mutex_t* pseudo_file_mutex() {
  static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
  return &mutex;
}

static int32_t open_pseudo_file(const void* content, size_t content_size) {
  pthread_mutex_lock(pseudo_file_mutex());

  int index = 0;
  int found = 0;
  for (int i = 0; i < sizeof(pseudo_files) / sizeof(pseudo_file); i++) {
    if (pseudo_files[i].fd == 0) {
      index = i;
      found = 1;
      break;
    }
  }
  if (!found) {
    // Too many pseudo files are opened.
    pthread_mutex_unlock(pseudo_file_mutex());
    return -1;
  }
  int32_t fd = index + kFDOffset;
  pseudo_files[index].content = content;
  pseudo_files[index].content_size = content_size;
  pseudo_files[index].current = 0;
  pseudo_files[index].fd = fd;

  pthread_mutex_unlock(pseudo_file_mutex());
  return fd;
}

static size_t read_pseudo_file(int32_t fd, void *p, int32_t n) {
  pthread_mutex_lock(pseudo_file_mutex());

  int32_t index = fd - kFDOffset;
  pseudo_file *file = &pseudo_files[index];
  size_t rest = file->content_size - file->current;
  if (rest < n) {
    n = rest;
  }
  memcpy(p, file->content + file->current, n);
  pseudo_files[index].current += n;

  pthread_mutex_unlock(pseudo_file_mutex());
  return n;
}

static void close_pseudo_file(int32_t fd) {
  pthread_mutex_lock(pseudo_file_mutex());

  int32_t index = fd - kFDOffset;
  pseudo_files[index].content = NULL;
  pseudo_files[index].content_size = 0;
  pseudo_files[index].current = 0;
  pseudo_files[index].fd = 0;

  pthread_mutex_unlock(pseudo_file_mutex());
}

int32_t hitsumabushi_closefd(int32_t fd) {
  if (fd >= kFDOffset) {
    close_pseudo_file(fd);
    return 0;
  }
  fprintf(stderr, "syscall close(%d) is not implemented\n", fd);
  return 0;
}

int32_t hitsumabushi_open(char *name, int32_t mode, int32_t perm) {
  if (strcmp(name, "/proc/self/auxv") == 0) {
    static const char auxv[] =
      "\x06\x00\x00\x00\x00\x00\x00\x00"  // _AT_PAGESZ tag (6)
      "\x00\x10\x00\x00\x00\x00\x00\x00"  // 4096 bytes per page
      "\x00\x00\x00\x00\x00\x00\x00\x00"  // Dummy bytes
      "\x00\x00\x00\x00\x00\x00\x00\x00"; // Dummy bytes
    return open_pseudo_file(auxv, sizeof(auxv) / sizeof(char));
  }
  if (strcmp(name, "/sys/kernel/mm/transparent_hugepage/hpage_pmd_size") == 0) {
    static const char hpage_pmd_size[] =
      "\x30\x5c"; // '0', '\n'
    return open_pseudo_file(hpage_pmd_size, sizeof(hpage_pmd_size) / sizeof(char));
  }
  fprintf(stderr, "syscall open(%s, %d, %d) is not implemented\n", name, mode, perm);
  const static int kENOENT = 0x2;
  return kENOENT;
}

int32_t hitsumabushi_read(int32_t fd, void *p, int32_t n) {
  if (fd >= kFDOffset) {
    return read_pseudo_file(fd, p, n);
  }
  fprintf(stderr, "syscall read(%d, %p, %d) is not implemented\n", fd, p, n);
  const static int kEBADF = 0x9;
  return kEBADF;
}

int32_t hitsumabushi_write1(uintptr_t fd, void *p, int32_t n) {
  static pthread_mutex_t m = PTHREAD_MUTEX_INITIALIZER;
  int32_t ret = 0;
  pthread_mutex_lock(&m);
  switch (fd) {
  case 1:
    ret = fwrite(p, 1, n, stdout);
    fflush(stdout);
    break;
  case 2:
    ret = fwrite(p, 1, n, stderr);
    fflush(stderr);
    break;
  default:
    fprintf(stderr, "syscall write(%lu, %p, %d) is not implemented\n", fd, p, n);
    ret = -EBADF;
    break;
  }
  pthread_mutex_unlock(&m);
  return ret;
}

int32_t hitsumabushi_lseek(uintptr_t fd, off_t offset, int32_t whence) {
  fprintf(stderr, "syscall lseek(%lu, %lu, %d) is not implemented\n", fd, offset, whence);
  return -ENOSYS;
}

int32_t hitsumabushi_fcntl(int32_t fd, int32_t cmd, int32_t arg)
{
  if (fd == 0 || fd == 1 || fd == 2) {
    if (cmd == F_GETFL) {
      return 0;
    }
  }
  fprintf(stderr, "syscall fcntl(%d, %d, %d) is not implemented\n", fd, cmd, arg);
  return -EBADF;
}

int32_t hitsumabushi_fstat(int32_t fd, struct stat *stat)
{
  fprintf(stderr, "syscall fstat(%d, %p) is not implemented\n", fd, stat);
  return -ENOSYS;
}

int32_t hitsumabushi_renameat(int32_t fd1, char* name1, int32_t fd2, char* name2)
{
  fprintf(stderr, "syscall renameat(%d, %s, %d, %s) is not implemented\n", fd1, name1, fd2, name2);
  return -ENOSYS;
}

int32_t hitsumabushi_fstatat(int32_t fd, char* name, struct stat* p, int32_t flags)
{
  fprintf(stderr, "syscall fstatat(%d, %s, %p, %d) is not implemented\n", fd, name, p, flags);
  return -ENOSYS;
}
