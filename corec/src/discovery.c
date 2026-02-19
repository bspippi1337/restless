#include "restless/core.h"
#include "restless/util.h"
#include <stdlib.h>

int restless_discover(const restless_discover_options* opt, restless_discover_result* out) {
  if (!opt || !out || !opt->domain) return 2;

  // Skeleton output: keep JSON stable so the Go side can call this later.
  // Future: implement evidence-based discovery, OpenAPI parsing, seed-only fuzzing, verification.
  out->json = str_fmt(
    "{\n"
    "  \"domain\": \"%s\",\n"
    "  \"status\": \"stub\",\n"
    "  \"budgets\": {\"seconds\": %d, \"pages\": %d},\n"
    "  \"verify\": %s,\n"
    "  \"fuzz\": %s,\n"
    "  \"endpoints\": [],\n"
    "  \"notes\": \"C core skeleton: discovery engine not implemented yet.\"\n"
    "}\n",
    opt->domain,
    opt->budget_seconds,
    opt->budget_pages,
    opt->verify ? "true" : "false",
    opt->fuzz ? "true" : "false"
  );
  return out->json ? 0 : 3;
}

void restless_free_result(restless_discover_result* r) {
  if (!r) return;
  free(r->json);
  r->json = NULL;
}
