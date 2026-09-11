#include <napi/native_api.h>
#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#define CHECK(expr) do { napi_status s=(expr); if(s!=napi_ok) { fprintf(stderr,"%s: status %d, line %d\n",#expr,s,__LINE__); exit(1); } } while(0)
static int completed;
static napi_value resolved(napi_env env, napi_callback_info info) {
    napi_value result, unused;
    size_t argc=1;
    int64_t total;
    CHECK(napi_get_cb_info(env,info,&argc,&result,NULL,NULL));
    CHECK(napi_get_value_int64(env,result,&total));
    if(total!=50005000) { fprintf(stderr,"bad async result: %lld\n",(long long)total); exit(2); }
    completed++;
    CHECK(napi_get_undefined(env,&unused));
    return unused;
}
static napi_value rejected(napi_env env, napi_callback_info info) {
    fprintf(stderr,"unexpected Promise rejection\n"); exit(3);
}
int main(int argc, char **argv) {
    if(argc!=2) return 2;
    napi_env env = NULL;
    CHECK(napi_create_ark_runtime(&env));
    napi_handle_scope scope;
    CHECK(napi_open_handle_scope(env,&scope));
    void *lib=dlopen(argv[1],RTLD_NOW|RTLD_LOCAL);
    if(!lib) { fprintf(stderr,"dlopen: %s\n",dlerror()); return 4; }
    napi_value (*init)(napi_env,napi_value)=dlsym(lib,"GoNapiInit");
    if(!init) return 5;
    napi_value exports, fn, args[2], result;
    CHECK(napi_create_object(env,&exports));
    if(!init(env,exports)) return 6;
    CHECK(napi_get_named_property(env,exports,"add",&fn));
    CHECK(napi_create_int32(env,20,&args[0]));
    CHECK(napi_create_int32(env,22,&args[1]));
    CHECK(napi_call_function(env,exports,fn,2,args,&result));
    int64_t number;
    CHECK(napi_get_value_int64(env,result,&number));
    if(number!=42) return 7;
    // Validate a type error can be cleared and the VM remains usable.
    CHECK(napi_create_string_utf8(env,"invalid",NAPI_AUTO_LENGTH,&args[0]));
    if(napi_call_function(env,exports,fn,2,args,&result)!=napi_pending_exception) return 8;
    CHECK(napi_get_and_clear_last_exception(env,&result));
    CHECK(napi_get_named_property(env,exports,"sumAsync",&fn));
    CHECK(napi_create_int32(env,10000,&args[0]));
    for(int i=0;i<16;i++) {
        napi_value promise, then, handlers[2];
        CHECK(napi_call_function(env,exports,fn,1,args,&promise));
        CHECK(napi_get_named_property(env,promise,"then",&then));
        CHECK(napi_create_function(env,"resolved",NAPI_AUTO_LENGTH,resolved,NULL,&handlers[0]));
        CHECK(napi_create_function(env,"rejected",NAPI_AUTO_LENGTH,rejected,NULL,&handlers[1]));
        CHECK(napi_call_function(env,promise,then,2,handlers,&result));
    }
    for(int i=0;completed<16 && i<1000;i++) {
        CHECK(napi_run_event_loop(env,napi_event_mode_nowait));
        usleep(10000);
    }
    if(completed!=16) { fprintf(stderr,"only %d promises resolved\n",completed); return 9; }
    CHECK(napi_close_handle_scope(env,scope));
    CHECK(napi_destroy_ark_runtime(&env));
    puts("PASS Ark runtime -> N-API -> Go; type errors, 16 async Promises and GC");
    // Go shared libraries must remain loaded until process exit.
    return 0;
}
