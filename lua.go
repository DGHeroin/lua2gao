package lua

/*
#cgo CFLAGS: -I ${SRCDIR}/inc
#cgo CFLAGS: -I ${SRCDIR}/lua
#cgo windows,!llua LDFLAGS: -L${SRCDIR}/libs/windows -llua -lm -lws2_32
#cgo linux,!llua LDFLAGS: -L${SRCDIR}/libs/linux -llua -lm
#cgo darwin,!llua LDFLAGS: -L${SRCDIR}/libs/macos -llua -lm

#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"

*/
import "C"
import (
    "bytes"
    "fmt"
    "log"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

//export gCallbackStr
func gCallbackStr(ref C.int, size C.ulonglong, str *C.char) {
    fn := lookup(int(ref))
    if fn != nil {
        fn(size, str)
    }
}

type (
    Context struct {
        C        chan []byte
        cmdQueue chan *luaCmd
        closeOnce sync.Once
        closeCh    chan struct{}
    }
)

func (c *Context) PushLuaMessage(val []byte) {
    c.cmdQueue <- &luaCmd{
        what: cmdLuaMessage,
        data: val,
    }
}

func (c *Context) DoString(code string) {
    c.cmdQueue <- &luaCmd{
        what: cmdDoString,
        data: []byte(code),
    }
}
func (c *Context) DoFile(filename string) {
    c.cmdQueue <- &luaCmd{
        what: cmdDoFile,
        data: []byte(filename),
    }
}

func (c *Context) Close() {
    c.closeOnce.Do(func() {
        close(c.closeCh)
        close(c.cmdQueue)
        close(c.C)
    })
}

const (
    cmdGoMessage = iota
    cmdLuaMessage
    cmdDoString
    cmdDoFile
)
type luaCmd struct {
    what int // 0
    data []byte
}
var (
    qps = int32(0)
    lastTime = time.Now()
)

func New() *Context {
    ctx := &Context{
        C:        make(chan []byte, 8),
        cmdQueue: make(chan *luaCmd, 8),
    }

    L := C.cInitLuaState()
    recvLuaMsg := func(size C.ulonglong, msg *C.char) {
        str := C.GoStringN(msg, C.int(size))
        ctx.C <- []byte(str)
    }
    ref := register(recvLuaMsg)

    initCode := fmt.Sprintf(`
local ref = %d

local hideOnGoMessage = _onGoMessage
_onGoMessage = nil
function PushGoMessage(msg)
    hideOnGoMessage(ref, msg)
end
function OnLuaMessage(msg)
    --print('OnLuaMessage need overwrite')
end

`, ref)
    ctx.DoString(initCode)

    go func() {
        defer func() {
            unregister(ref)
            close(ctx.C)
            close(ctx.closeCh)
        }()

        for {
            select {
            case <-ctx.closeCh:
                return
            case cmd := <-ctx.cmdQueue:
                if cmd == nil {
                    return
                }
                newQPS := atomic.AddInt32(&qps, 1)
                if time.Now().Sub(lastTime) > time.Second {
                    log.Println("qps:", newQPS)
                    atomic.StoreInt32(&qps, 0)
                    lastTime = time.Now()
                }
                switch cmd.what {
                case cmdDoString:
                    code2 := C.CString(string(cmd.data))

                    C.cDoString(L, code2)

                    C.free(unsafe.Pointer(code2))
                case cmdDoFile:
                    code2 := C.CString(string(cmd.data))

                    C.cDoFile(L, code2)

                    C.free(unsafe.Pointer(code2))
                case cmdLuaMessage:
                    str := C.CString(bytes.NewBuffer(cmd.data).String())

                    size := len(cmd.data)
                    C.cOnLuaMessage(L, C.ulonglong(size), str)

                    C.free(unsafe.Pointer(str))
                }

            }
        }
    }()
    return ctx
}

//export GetTimeNow
func GetTimeNow() int64 {
    return time.Now().UnixNano()
}
