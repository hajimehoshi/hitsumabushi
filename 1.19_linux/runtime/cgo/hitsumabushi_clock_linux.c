#include <time.h>

int hitsumabushi_clock_gettime(clockid_t clk_id, struct timespec *tp) {
  return clock_gettime(clk_id, tp);
}
