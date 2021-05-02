#include "clua.h"
extern void gCallbackStr(int ref, size_t sz, const char* str);
static int onGoMessage(lua_State* L) {
    int         ref;
    size_t      size;
    const char* str;
    ref = luaL_checkinteger(L, 1);
    str = luaL_checklstring(L, 2, &size);
    gCallbackStr(ref, size, str);
    return 0;
}
static int lGetTimeNow(lua_State* L) {
    lua_pushnumber(L, GetTimeNow());
    return 1;
}

void c_register_lib(lua_State* L, const char* name, lua_CFunction fn) {
    luaL_requiref(L, name, fn, 1);
    lua_pop(L, 1);
}

void* cInitLuaState() {
    lua_State* L = luaL_newstate();
    luaL_openlibs(L);

    lua_pushcfunction(L, onGoMessage);
    lua_setglobal(L, "_onGoMessage");

    lua_pushcfunction(L, lGetTimeNow);
    lua_setglobal(L, "GetTimeNow");

    // register libs
    c_register_lib(L, "serialize", luaopen_serialize);
    c_register_lib(L, "cmsgpack", luaopen_cmsgpack);
    c_register_lib(L, "pb", luaopen_pb);
    fflush(stdout);
    return L;
}

 void cOnLuaMessage(void* ptr, size_t sz, char* msg) {
    lua_State* L = (lua_State*)ptr;

    lua_getglobal(L, "OnLuaMessage");
    if(lua_isfunction(L, -1)) {
        lua_pushlstring(L, msg, sz);
        if (0 != lua_pcall(L, 1, 0, 0)) {
            printf("err:%s\n", luaL_checkstring(L, -1));
        }
    } else {
        printf("no on lua message");
    }
}

void* cDoString(void* ptr,char* code) {
    lua_State* L = (lua_State*)ptr;
    if (0 != luaL_dostring(L, code)) {
        printf("err:%s\n", luaL_checkstring(L, -1));
    }
    return NULL;
}

void* cDoFile(void* ptr,char* filepath) {
    lua_State* L = (lua_State*)ptr;
    if (0 != luaL_dofile(L, filepath)) {
        printf("err:%s\n", luaL_checkstring(L, -1));
    }
    return NULL;
}