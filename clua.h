#ifndef _LUA_
#define _LUA_

#include "inc.h"
#include "c-lib.h"

void* cInitLuaState();
void* cDoString(void* ptr, char* code);
void* cDoFile(void* ptr, char* code);

void cOnLuaMessage(void* ptr, size_t sz, char* msg);

extern long long GetTimeNow();
#endif