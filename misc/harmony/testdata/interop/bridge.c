#include <pthread.h>
#include <stdint.h>
#include <stdlib.h>
#include "_cgo_export.h"

static void *callback(void *arg) {
    return (void *)(intptr_t)GoCallback((int)(intptr_t)arg);
}

int callback_thread(int n) {
    pthread_t thread;
    void *result;
    if (pthread_create(&thread, NULL, callback, (void *)(intptr_t)n) != 0)
        abort();
    if (pthread_join(thread, &result) != 0)
        abort();
    return (int)(intptr_t)result;
}
