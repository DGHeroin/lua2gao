package main

import (
    "github.com/DGHeroin/lua2go"
    "log"
    "os"
    "sync"
)

func startInstance(wg *sync.WaitGroup)  {
    wg.Add(1)
    defer wg.Done()
    L := lua.New()
    L.DoString(`
function OnLuaMessage(msg)
    --print('lua 收到 go 消息:', msg)
end
PushGoMessage('hello world!')
`)
    go func() {
        for {
            //time.Sleep(time.Millisecond * 20)
            L.PushLuaMessage([]byte("你好"))
        }
    }()
    if len(os.Args) == 2 {
        L.DoFile(os.Args[1])
    }


    for {
        select {
        case msg := <-L.C:
            log.Printf("go 收到 lua 消息:[%s]", msg)
        }
    }
}

func main() {
    wg := &sync.WaitGroup{}

    for i := 0; i < 5; i++ {
        go startInstance(wg)
    }

    wg.Wait()
}
