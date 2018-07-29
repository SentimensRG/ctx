# refctx

Package `refctx` provides contextualized reference counting. It exports a context
whose lifetime is bound to a `RefCtr` instance.  The `RefCtr` keeps track of how
many references to a context are held and cancels this context when the refcount
reaches zero.

`refctx` works similarly to `sync.WaitGroup`.

## Examples

```go
package main

import "github.com/SentimensRG/ctx/refctx"

func main() {
    ctx, ctr := refctx.WithRefCount(context.Background())

    for i := 0; i < 5; i++ {
        ctr.Incr()
        go func() {
            defer ctr.Decr()

            time.Sleep(time.Second * i)
        }()
    }

    <-ctx.Done()  // fires when refcount falls back to zero
}

```

A common use-case for `refctx` is to manage timeouts.  Consider the following
example using `github.com/gorilla/websocket`.

```go
import (
    "time"

    "github.com/SentimensRG/ctx/refctx"
    "github.com/SentimensRG/ctx/sigctx"

    "github.com/gorilla/websocket"
)

const (
    pingDeadline = time.Second * 1
    pongDeadline = pingDeadline * 2
)

func main() {

    conn := openWebsocketConnection()

    ctx, ctr := refctx.WithRefCount(sigctx.New())  // good place for sigctx
    rc.Incr()  // start with one refcount

    go func() {
        for range time.NewTicker(heartbeatInterval).C {
            select {
            case <-c.Done():
                // c.Done fires either when the process receives an OS signal, or
                // when the peer took too long to respond to a ping.
                return
            default:
                deadline := time.Now().Add(pingDeadline)
                _ = conn.WriteControl(websocket.PingMessage, nil, deadline)
                go func() {
                    <-time.After(pongDeadline)
                    rc.Decr()
                }()
            }
        }
    }()

    conn.SetPongHandler(func(_ string) (_ error) {
        rc.Incr()
        return
    })

    businessLogic(c, conn)
    <-c.Done()
}
```