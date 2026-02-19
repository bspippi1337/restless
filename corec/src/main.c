#include "restless/core.h"
#include "restless/util.h"
#include <stdio.h>
#include <stdlib.h>

int restless_doctor(void);

static void usage(void) {
  printf("restless-core (C skeleton)\n\n");
  printf("Usage:\n");
  printf("  restless-core doctor\n");
  printf("  restless-core discover <domain> [--verify] [--fuzz] [--budget-seconds N] [--budget-pages N]\n");
  printf("\n");
}

static int arg_has(const char* a, const char* name) { return str_eq(a, name); }

int main(int argc, char** argv) {
  if (argc < 2 || arg_has(argv[1], "--help") || arg_has(argv[1], "-h")) {
    usage();
    return 0;
  }

  if (arg_has(argv[1], "doctor")) {
    return restless_doctor();
  }

  if (arg_has(argv[1], "discover")) {
    if (argc < 3) {
      fprintf(stderr, "discover requires <domain>\n");
      return 2;
    }
    restless_discover_options opt = {
      .domain = argv[2],
      .budget_seconds = 20,
      .budget_pages = 8,
      .verify = 0,
      .fuzz = 0
    };

    for (int i = 3; i < argc; i++) {
      const char* a = argv[i];
      if (arg_has(a, "--verify")) opt.verify = 1;
      else if (arg_has(a, "--fuzz")) opt.fuzz = 1;
      else if (arg_has(a, "--budget-seconds") && i + 1 < argc) {
        opt.budget_seconds = atoi(argv[++i]);
      } else if (arg_has(a, "--budget-pages") && i + 1 < argc) {
        opt.budget_pages = atoi(argv[++i]);
      }
    }

    restless_discover_result r = {0};
    int rc = restless_discover(&opt, &r);
    if (rc != 0) {
      fprintf(stderr, "discover failed: %d\n", rc);
      return rc;
    }
    fputs(r.json, stdout);
    restless_free_result(&r);
    return 0;
  }

  fprintf(stderr, "unknown command: %s\n", argv[1]);
  usage();
  return 2;
}
