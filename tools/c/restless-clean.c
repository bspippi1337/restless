/*
 * RESTLESS repo cleaner. Portable C11 with POSIX/Windows backends.
 * Dry-run is default. This replaces the old scripts/cleanup.sh.
 * POSIX: cc -std=c11 -D_DEFAULT_SOURCE -Wall -Wextra -Werror -o restless-clean tools/c/restless-clean.c
 * MSVC: cl /W4 /TC tools\c\restless-clean.c
 */
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/stat.h>
#ifdef _WIN32
#include <windows.h>
#include <direct.h>
#define CHDIR _chdir
#define SEP '\\'
#else
#include <unistd.h>
#include <dirent.h>
#define CHDIR chdir
#define SEP '/'
#endif

#define MAX_PATH_BYTES 8192
static int apply = 0, quiet = 0;
static unsigned long files_seen = 0, dirs_seen = 0;

static int join_path(char *dest, size_t cap, const char *parent, const char *name) {
    int n = snprintf(dest, cap, "%s%c%s", parent, SEP, name);
    if (n < 0 || (size_t)n >= cap) {
        fprintf(stderr, "Path exceeds safety limit.\n");
        return -1;
    }
    return 0;
}

/* Never follow symlinks, junctions, or reparse points. */
static int kind_of(const char *path) {
#ifdef _WIN32
    DWORD flags = GetFileAttributesA(path);
    if (flags == INVALID_FILE_ATTRIBUTES) return 0;
    if (flags & FILE_ATTRIBUTE_REPARSE_POINT) return 3;
    return (flags & FILE_ATTRIBUTE_DIRECTORY) ? 2 : 1;
#else
    struct stat st;
    if (lstat(path, &st) != 0) return (errno == ENOENT) ? 0 : -1;
    if (S_ISLNK(st.st_mode)) return 3;
    return S_ISDIR(st.st_mode) ? 2 : (S_ISREG(st.st_mode) ? 1 : 3);
#endif
}

static int remove_file(const char *path) {
    ++files_seen;
    if (!quiet) printf("%s file: %s\n", apply ? "Remove" : "Would remove", path);
    if (!apply) return 0;
#ifdef _WIN32
    if (DeleteFileA(path)) return 0;
#else
    if (remove(path) == 0) return 0;
#endif
    fprintf(stderr, "Failed to delete file: %s\n", path);
    return -1;
}

static int remove_directory(const char *path) {
    ++dirs_seen;
    if (!quiet) printf("%s directory: %s\n", apply ? "Remove" : "Would remove", path);
    if (!apply) return 0;
#ifdef _WIN32
    if (RemoveDirectoryA(path)) return 0;
#else
    if (rmdir(path) == 0) return 0;
#endif
    fprintf(stderr, "Failed to remove directory: %s\n", path);
    return -1;
}

static int clean_tree(const char *path, unsigned int depth) {
    int kind = kind_of(path);
    if (kind == 0) return 0;
    if (kind < 0) { fprintf(stderr, "Unable to inspect %s\n", path); return -1; }
    if (kind == 3) { fprintf(stderr, "Refusing symlink or special file: %s\n", path); return -1; }
    if (kind == 1) return remove_file(path);
    if (depth > 64) { fprintf(stderr, "Directory depth limit exceeded: %s\n", path); return -1; }
#ifdef _WIN32
    char pattern[MAX_PATH_BYTES];
    if (join_path(pattern, sizeof pattern, path, "*") != 0) return -1;
    WIN32_FIND_DATAA item;
    HANDLE find = FindFirstFileA(pattern, &item);
    if (find != INVALID_HANDLE_VALUE) {
        do {
            if (strcmp(item.cFileName, ".") == 0 || strcmp(item.cFileName, "..") == 0) continue;
            char child[MAX_PATH_BYTES];
            if (join_path(child, sizeof child, path, item.cFileName) != 0 || clean_tree(child, depth+1) != 0) {
                FindClose(find);
                return -1;
            }
        } while (FindNextFileA(find, &item));
        FindClose(find);
    }
#else
    DIR *dir = opendir(path);
    if (!dir) { fprintf(stderr, "Cannot open directory: %s\n", path); return -1; }
    struct dirent *item;
    while ((item = readdir(dir)) != NULL) {
        if (strcmp(item->d_name, ".") == 0 || strcmp(item->d_name, "..") == 0) continue;
        char child[MAX_PATH_BYTES];
        if (join_path(child, sizeof child, path, item->d_name) != 0 || clean_tree(child, depth+1) != 0) {
            closedir(dir);
            return -1;
        }
    }
    closedir(dir);
#endif
    return remove_directory(path);
}

static int is_repo_root(void) {
    FILE *mod = fopen("go.mod", "rb");
    if (!mod) return 0;
    fclose(mod);
    return kind_of(".git") == 2 || kind_of(".git") == 1;
}

int main(int argc, char **argv) {
    const char *root = ".";
    for (int i = 1; i < argc; ++i) {
        if (strcmp(argv[i], "--apply") == 0) apply = 1;
        else if (strcmp(argv[i], "--dry-run") == 0) apply = 0;
        else if (strcmp(argv[i], "--root") == 0 && i + 1 < argc) root = argv[++i];
        else if (strcmp(argv[i], "--help") == 0) {
            puts("RESTLESS cleaner: --root DIRECTORY [--dry-run|--apply]. Default: dry-run.\nOnly generated build/, dist/, bin/, logs/, and .fixall-logs/ are considered.\nSymlinks and special files stop processing. Git and source files are untouched.");
            return 0;
        } else {
            fprintf(stderr, "Unknown argument: %s\n", argv[i]);
            return 2;
        }
    }
    if (CHDIR(root) != 0 || !is_repo_root()) {
        fprintf(stderr, "Refusing cleanup: --root must contain go.mod and .git.\n");
        return 2;
    }
    const char *targets[] = {"build", "dist", "bin", "logs", ".fixall-logs"};
    printf("RESTLESS cleaner (%s), scoped to repository root.\n", apply ? "APPLY" : "DRY RUN");

    /* Validate every top-level target before deleting anything. */
    for (size_t i = 0; i < sizeof(targets)/sizeof(targets[0]); ++i) {
        int kind = kind_of(targets[i]);
        if (kind < 0 || kind == 3 || kind == 1) {
            fprintf(stderr, "Unsafe cleanup target: %s\n", targets[i]);
            return 3;
        }
    }
    if (apply) {
        /* Dry preflight of every nested entry, before the first deletion. */
        apply = 0; quiet = 1;
        for (size_t i = 0; i < sizeof(targets)/sizeof(targets[0]); ++i) {
            if (clean_tree(targets[i], 0) != 0) return 4;
        }
        files_seen = 0; dirs_seen = 0; quiet = 0; apply = 1;
    }
    for (size_t i = 0; i < sizeof(targets)/sizeof(targets[0]); ++i) {
        if (clean_tree(targets[i], 0) != 0) return 4;
    }
    printf("Done. %lu files and %lu directories %s.\n", files_seen, dirs_seen, apply ? "removed" : "eligible");
    return 0;
}
