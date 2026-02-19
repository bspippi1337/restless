#include <stdio.h>
#include <stdlib.h>

int restless_doctor(void) {
  printf("{\n");
  printf("  \"name\": \"restless-core\",\n");
  printf("  \"version\": \"0.0.0-skeleton\",\n");
  printf("  \"notes\": \"C core skeleton (no external deps). HTTPS backend not yet wired.\",\n");
#if defined(_WIN32)
  printf("  \"platform\": \"windows\",\n");
#elif defined(__APPLE__)
  printf("  \"platform\": \"macos\",\n");
#elif defined(__ANDROID__)
  printf("  \"platform\": \"android\",\n");
#elif defined(__linux__)
  printf("  \"platform\": \"linux\",\n");
#else
  printf("  \"platform\": \"unknown\",\n");
#endif
  printf("  \"env\": {\n");
  printf("    \"COLUMNS\": \"%s\"\n", getenv("COLUMNS") ? getenv("COLUMNS") : "");
  printf("  }\n");
  printf("}\n");
  return 0;
}
