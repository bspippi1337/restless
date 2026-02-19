#include "restless/util.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdarg.h>

int str_eq(const char* a, const char* b) {
  if (!a || !b) return 0;
  return strcmp(a, b) == 0;
}

char* xstrdup(const char* s) {
  if (!s) return NULL;
  size_t n = strlen(s);
  char* out = (char*)malloc(n + 1);
  if (!out) return NULL;
  memcpy(out, s, n + 1);
  return out;
}

char* str_fmt(const char* fmt, ...) {
  va_list ap;
  va_start(ap, fmt);
  va_list ap2;
  va_copy(ap2, ap);
  int need = vsnprintf(NULL, 0, fmt, ap);
  va_end(ap);
  if (need < 0) { va_end(ap2); return NULL; }
  char* buf = (char*)malloc((size_t)need + 1);
  if (!buf) { va_end(ap2); return NULL; }
  vsnprintf(buf, (size_t)need + 1, fmt, ap2);
  va_end(ap2);
  return buf;
}
