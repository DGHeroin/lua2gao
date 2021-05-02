package lua

import "C"
import "sync"

var mu sync.Mutex
var index int
var fns = make(map[int]Callback)

type Callback func(C.ulonglong, *C.char)

func register(fn Callback) int {
    mu.Lock()
    defer mu.Unlock()
    index++
    for fns[index] != nil {
        index++
    }
    fns[index] = fn
    return index
}

func lookup(i int) Callback{
    mu.Lock()
    defer mu.Unlock()
    return fns[i]
}

func unregister(i int) {
    mu.Lock()
    defer mu.Unlock()
    delete(fns, i)
}
