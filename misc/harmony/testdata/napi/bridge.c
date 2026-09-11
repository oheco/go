#include <napi/native_api.h>
#include <stdlib.h>
#include "_cgo_export.h"

static napi_value fail(napi_env env, const char *message) {
    napi_throw_error(env, NULL, message);
    return NULL;
}

static int integer(napi_env env, napi_value value, int32_t *out) {
    double d;
    if (napi_get_value_double(env, value, &d) != napi_ok ||
        !(d >= INT32_MIN && d <= INT32_MAX)) return 0;
    *out = (int32_t)d;
    return d == *out;
}

static napi_value add(napi_env env, napi_callback_info info) {
    size_t argc = 2;
    napi_value args[2], result;
    int32_t a, b;
    if (napi_get_cb_info(env, info, &argc, args, NULL, NULL) != napi_ok ||
        argc != 2 || !integer(env, args[0], &a) || !integer(env, args[1], &b))
        return fail(env, "add expects two signed 32-bit integers");
    if (napi_create_int64(env, GoAdd(a, b), &result) != napi_ok)
        return fail(env, "cannot create result");
    return result;
}

typedef struct {
    napi_async_work work;
    napi_deferred deferred;
    int32_t n;
    int64_t result;
} sum_work;

static void execute_sum(napi_env env, void *data) {
    sum_work *w = data;
    w->result = GoSum(w->n);
}

static void complete_sum(napi_env env, napi_status status, void *data) {
    sum_work *w = data;
    napi_value result;
    if (status == napi_ok && napi_create_int64(env, w->result, &result) == napi_ok) {
        napi_resolve_deferred(env, w->deferred, result);
    } else {
        napi_value message;
        napi_create_string_utf8(env, "Go async work failed", NAPI_AUTO_LENGTH, &message);
        napi_create_error(env, NULL, message, &result);
        napi_reject_deferred(env, w->deferred, result);
    }
    napi_delete_async_work(env, w->work);
    free(w);
}

static napi_value sum_async(napi_env env, napi_callback_info info) {
    size_t argc = 1;
    napi_value arg, promise, name;
    int32_t n;
    if (napi_get_cb_info(env, info, &argc, &arg, NULL, NULL) != napi_ok ||
        argc != 1 || !integer(env, arg, &n) || n < 0 || n > 1000000)
        return fail(env, "sumAsync expects an integer between 0 and 1000000");
    sum_work *w = calloc(1, sizeof(*w));
    if (!w) return fail(env, "cannot allocate async work");
    w->n = n;
    if (napi_create_promise(env, &w->deferred, &promise) != napi_ok ||
        napi_create_string_utf8(env, "GoSum", NAPI_AUTO_LENGTH, &name) != napi_ok ||
        napi_create_async_work(env, NULL, name, execute_sum, complete_sum, w, &w->work) != napi_ok ||
        napi_queue_async_work(env, w->work) != napi_ok) {
        if (w->work) napi_delete_async_work(env, w->work);
        free(w);
        return fail(env, "cannot queue Go async work");
    }
    return promise;
}

__attribute__((visibility("default")))
napi_value GoNapiInit(napi_env env, napi_value exports) {
    napi_property_descriptor properties[] = {
        {"add", NULL, add, NULL, NULL, NULL, napi_default, NULL},
        {"sumAsync", NULL, sum_async, NULL, NULL, NULL, napi_default, NULL}
    };
    if (napi_define_properties(env, exports, 2, properties) != napi_ok)
        return fail(env, "cannot register Go exports");
    return exports;
}

static napi_module module = {
    .nm_version = 1,
    .nm_register_func = GoNapiInit,
    .nm_modname = "goharmony",
};

__attribute__((constructor)) static void register_module(void) {
    napi_module_register(&module);
}
