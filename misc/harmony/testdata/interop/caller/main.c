#include <dlfcn.h>
#include <pthread.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static int (*call_go)(int);
static int (*check_environment)(void);
static int (*check_args)(int, char **);
#ifdef STATIC_GO
extern int GoCallback(int);
extern int GoCheckEnvironment(void);
extern int GoCheckArgs(int, char **);
#endif

static void *thread_main(void *arg) {
    int n = (int)(intptr_t)arg;
    return (void *)(intptr_t)(call_go(n) != n * 3);
}

int main(int argc, char **argv) {
#ifdef STATIC_GO
    call_go = GoCallback;
    check_environment = GoCheckEnvironment;
    check_args = GoCheckArgs;
#else
    if (argc < 2) return 2;
    void *lib = dlopen(argv[1], RTLD_NOW | RTLD_LOCAL);
    if (!lib) { fprintf(stderr, "%s\n", dlerror()); return 3; }
    call_go = (int (*)(int))dlsym(lib, "GoCallback");
    if (!call_go) { fprintf(stderr, "%s\n", dlerror()); return 4; }
    check_environment = (int (*)(void))dlsym(lib, "GoCheckEnvironment");
    check_args = (int (*)(int, char **))dlsym(lib, "GoCheckArgs");
#endif
    if (getenv("OHOS_GO_INTEROP") && (!check_environment || check_environment() != 1)) {
        fprintf(stderr, "Go library did not inherit the C process environment\n");
        return 7;
    }
    if (!check_args || check_args(argc, argv) != 1) {
        fprintf(stderr, "Go library did not inherit the C process arguments\n");
        return 8;
    }
    pthread_t threads[16];
    for (int i = 0; i < 16; i++)
        if (pthread_create(&threads[i], NULL, thread_main, (void *)(intptr_t)i)) return 5;
    for (int i = 0; i < 16; i++) {
        void *result;
        if (pthread_join(threads[i], &result) || result) return 6;
    }
    puts("PASS C host -> Go library, arguments, concurrent C threads and GC");
    return 0;
}
