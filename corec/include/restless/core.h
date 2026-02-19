#pragma once
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
  const char* domain;     // input domain
  int budget_seconds;     // time budget
  int budget_pages;       // crawl budget
  int verify;             // 0/1
  int fuzz;               // 0/1
} restless_discover_options;

typedef struct {
  char* json;             // heap-allocated JSON output
} restless_discover_result;

int restless_discover(const restless_discover_options* opt, restless_discover_result* out);
void restless_free_result(restless_discover_result* r);

#ifdef __cplusplus
}
#endif
